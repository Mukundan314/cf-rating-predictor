package cache

import (
	"sync"
	"time"

	"github.com/cheran-senthil/cf-rating-predictor/calculator"
	"github.com/mukundan314/go-codeforces"
	"github.com/sirupsen/logrus"
)

type Cache struct {
	updateRatingBefore time.Duration
	lastUpdateTime     time.Time

	userRating map[string]int
	ratingLock sync.RWMutex

	ratingChanges map[int][]codeforces.RatingChange
	changesLock   sync.RWMutex
}

func NewCache(updateRatingBefore time.Duration) *Cache {
	return &Cache{
		userRating:         make(map[string]int),
		ratingChanges:      make(map[int][]codeforces.RatingChange),
		updateRatingBefore: updateRatingBefore,
	}
}

func (c *Cache) GetRating(handle string) int {
	c.ratingLock.RLock()
	defer c.ratingLock.RUnlock()

	if v, ok := c.userRating[handle]; ok {
		return v
	} else {
		return 1500
	}
}

func (c *Cache) GetRatingChanges(contestID int) []codeforces.RatingChange {
	c.ratingLock.RLock()
	defer c.ratingLock.RUnlock()

	return c.ratingChanges[contestID]
}

func (c *Cache) UpdateContestRatingChanges(contestID int) error {
	logrus.WithField("contestID", contestID).Debug("Recomputing Rating Changes")

	_, _, ranklistRows, err := codeforces.GetContestStandings(contestID, 1, 0, nil, 0, false)
	if err != nil {
		return err
	}

	userRating := make(map[string]int)

	for _, ranklistRow := range ranklistRows {
		for _, member := range ranklistRow.Party.Members {
			userRating[member.Handle] = c.GetRating(member.Handle)
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
			OldRating:               c.GetRating(handle),
			NewRating:               c.GetRating(handle) + delta,
		})
	}

	c.changesLock.Lock()
	defer c.changesLock.Unlock()

	c.ratingChanges[contestID] = ratingChanges

	return nil
}

func (c *Cache) UpdateUserRatings() error {
	logrus.Debug("Updating User Ratings")

	users, err := codeforces.GetUserRatedList(false)
	if err != nil {
		return err
	}

	c.ratingLock.Lock()
	defer c.ratingLock.Unlock()

	for _, user := range users {
		c.userRating[user.Handle] = user.Rating
	}

	return nil
}

func (c *Cache) Update() error {
	contests, err := codeforces.GetContestList(false)
	if err != nil {
		return err
	}

	for _, contest := range contests {
		startTime := time.Unix(int64(contest.StartTimeSeconds), 0)

		if time.Until(startTime) < c.updateRatingBefore && startTime.Sub(c.lastUpdateTime) > c.updateRatingBefore {
			c.lastUpdateTime = time.Now()
			if err := c.UpdateUserRatings(); err != nil {
				return err
			}
		}

		if contest.Phase == "CODING" || contest.Phase == "PENDING_SYSTEM_TEST" {
			if err := c.UpdateContestRatingChanges(contest.ID); err != nil {
				return err
			}
		}
	}

	return err
}
