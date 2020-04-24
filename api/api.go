package api

import (
	"encoding/json"

	"github.com/mukundan314/go-codeforces"
)

type cache interface {
	GetRating(handle string) int
	GetRatingChanges(contestId int) []codeforces.RatingChange
}

type response struct {
	Status  string          `json:"status"`
	Comment string          `json:"comment,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}
