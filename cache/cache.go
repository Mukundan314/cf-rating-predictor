package cache

import (
	"sync"
	"time"

	"github.com/cheran-senthil/cf-rating-predictor/calculator"
	"github.com/mukundan314/go-codeforces"
	"github.com/sirupsen/logrus"
)

type Cache struct {
	lastUpdateTime time.Time

	userRating     map[string]int
	userRatingLock sync.RWMutex

	ratingChanges     map[int][]codeforces.RatingChange
	ratingChangesLock sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		userRating:    make(map[string]int),
		ratingChanges: make(map[int][]codeforces.RatingChange),
	}
}

func (c *Cache) GetRating(handle string) int {
	c.userRatingLock.RLock()
	defer c.userRatingLock.RUnlock()

	if v, ok := c.userRating[handle]; ok {
		return v
	} else {
		return 1400
	}
}

func (c *Cache) GetRatingChanges(contestID int) []codeforces.RatingChange {
	c.userRatingLock.RLock()
	defer c.userRatingLock.RUnlock()

	if c.ratingChanges[contestID] != nil {
		return c.ratingChanges[contestID]
	}

	return []codeforces.RatingChange{}
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

	c.ratingChangesLock.Lock()
	defer c.ratingChangesLock.Unlock()

	c.ratingChanges[contestID] = ratingChanges

	return nil
}

func (c *Cache) UpdateUserRatings() error {
	logrus.Debug("Updating User Ratings")

	c.lastUpdateTime = time.Now()

	users, err := codeforces.GetUserRatedList(false)
	if err != nil {
		return err
	}

	c.userRatingLock.Lock()
	defer c.userRatingLock.Unlock()

	for _, user := range users {
		c.userRating[user.Handle] = user.Rating
	}

	return nil
}

func (c *Cache) ClearContestRatingChanges(contestID int) {
	if c.ratingChanges[contestID] != nil {
		logrus.WithField("contestID", contestID).Debug("Clearing Rating Changes")

		c.ratingChangesLock.Lock()
		defer c.ratingChangesLock.Unlock()

		c.ratingChanges[contestID] = nil
	}
}

func shouldUpdateRating(
	contest codeforces.Contest,
	updateRatingBeforeContest time.Duration,
	lastUpdateTime time.Time,
) bool {
	startTime := time.Unix(int64(contest.StartTimeSeconds), 0)

	return (time.Now().After(startTime.Add(-updateRatingBeforeContest)) &&
		lastUpdateTime.Before(startTime.Add(-updateRatingBeforeContest)))
}

func shouldUpdateRatingChanges(
	contest codeforces.Contest,
	updateRatingChangesAfterContest time.Duration,
) bool {
	endTime := time.Unix(int64(contest.StartTimeSeconds+contest.DurationSeconds), 0)

	return (contest.Phase != "BEFORE" &&
		contest.Phase != "SYSTEM_TEST" &&
		(time.Now().Before(endTime.Add(updateRatingChangesAfterContest)) || contest.Phase != "FINISHED"))
}

func shouldClearRatingChanges(
	contest codeforces.Contest,
	clearRatingChangesAfterContest time.Duration,
) bool {
	endTime := time.Unix(int64(contest.StartTimeSeconds+contest.DurationSeconds), 0)
	return (contest.Phase == "FINISHED" && time.Now().After(endTime.Add(clearRatingChangesAfterContest)))
}

func (c *Cache) Update(
	updateRatingBeforeContest time.Duration,
	updateRatingChangesAfterContest time.Duration,
	clearRatingChangesAfterContest time.Duration,
) error {
	contests, err := codeforces.GetContestList(false)
	if err != nil {
		return err
	}

	for _, contest := range contests {
		if shouldUpdateRating(contest, updateRatingBeforeContest, c.lastUpdateTime) {
			if err := c.UpdateUserRatings(); err != nil {
				return err
			}
		}

		if shouldUpdateRatingChanges(contest, updateRatingChangesAfterContest) {
			if err := c.UpdateContestRatingChanges(contest.ID); err != nil {
				return err
			}
		}

		if shouldClearRatingChanges(contest, clearRatingChangesAfterContest) {
			c.ClearContestRatingChanges(contest.ID)
		}
	}

	return err
}
