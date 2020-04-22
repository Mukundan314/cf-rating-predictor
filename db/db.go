package db

import (
	"sync"
	"time"

	"github.com/cheran-senthil/cf-rating-predictor/calculator"
	"github.com/mukundan314/go-codeforces"
	"github.com/sirupsen/logrus"
)

type DB struct {
	userRating         map[string]int
	ratingChanges      map[int][]codeforces.RatingChange
	updateRatingBefore time.Duration
	lastUpdateTime     time.Time
	ratingLock         sync.RWMutex
	changesLock        sync.RWMutex
}

func NewDB(updateRatingBefore time.Duration) *DB {
	return &DB{
		userRating:         make(map[string]int),
		ratingChanges:      make(map[int][]codeforces.RatingChange),
		updateRatingBefore: updateRatingBefore,
	}
}

func (d *DB) GetRating(handle string) int {
	d.ratingLock.RLock()
	defer d.ratingLock.RUnlock()

	if v, ok := d.userRating[handle]; ok {
		return v
	} else {
		return 1500
	}
}

func (d *DB) GetRatingChanges(contestID int) []codeforces.RatingChange {
	d.ratingLock.RLock()
	defer d.ratingLock.RUnlock()

	return d.ratingChanges[contestID]
}

func (d *DB) UpdateContestRatingChanges(contestID int) error {
	logrus.WithField("contestID", contestID).Debug("Recomputing Rating Changes")

	_, _, ranklistRows, err := codeforces.GetContestStandings(contestID, 1, 0, nil, 0, false)
	if err != nil {
		return err
	}

	userRating := make(map[string]int)

	for _, ranklistRow := range ranklistRows {
		for _, member := range ranklistRow.Party.Members {
			userRating[member.Handle] = d.GetRating(member.Handle)
		}
	}

	ratingChanges := []codeforces.RatingChange{}
	ratingUpdateTimeSeconds := time.Now().Unix()

	deltas := calculator.CalculateRatingChanges(userRating, ranklistRows)
	for handle, delta := range deltas {
		ratingChanges = append(ratingChanges, codeforces.RatingChange{
			ContestID:               contestID,
			Handle:                  handle,
			RatingUpdateTimeSeconds: int(ratingUpdateTimeSeconds),
			OldRating:               d.GetRating(handle),
			NewRating:               d.GetRating(handle) + delta,
		})
	}

	d.changesLock.Lock()
	defer d.changesLock.Unlock()

	d.ratingChanges[contestID] = ratingChanges

	return nil
}

func (d *DB) UpdateUserRatings() error {
	logrus.Debug("Updating User Ratings")

	users, err := codeforces.GetUserRatedList(false)
	if err != nil {
		return err
	}

	d.ratingLock.Lock()
	defer d.ratingLock.Unlock()

	for _, user := range users {
		d.userRating[user.Handle] = user.Rating
	}

	return nil
}

func (d *DB) Update() error {
	logrus.Debug("Updating DB")

	contests, err := codeforces.GetContestList(false)
	if err != nil {
		return err
	}

	for _, contest := range contests {
		startTime := time.Unix(int64(contest.StartTimeSeconds), 0)

		if contest.Phase == "BEFORE" && time.Until(startTime) < d.updateRatingBefore && startTime.Sub(d.lastUpdateTime) > d.updateRatingBefore {
			d.lastUpdateTime = time.Now()
			if err := d.UpdateUserRatings(); err != nil {
				return err
			}
		}

		if contest.Phase == "CODING" || contest.Phase == "PENDING_SYSTEM_TEST" {
			if err := d.UpdateContestRatingChanges(contest.ID); err != nil {
				return err
			}
		}
	}

	return err
}
