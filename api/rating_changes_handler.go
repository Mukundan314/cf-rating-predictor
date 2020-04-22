package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"
)

type RatingChangesHandler struct {
	DB db
}

func (h RatingChangesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("request", r).Debug()

	w.Header().Add("Content-Type", "application/json")

	if r.Method != "GET" {
		res := response{
			Status:  "FAILED",
			Comment: "Methods other than GET are not supported",
		}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			logrus.WithError(err).Error()
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}

		return
	}

	unparsedContestID := r.URL.Query().Get("contestId")
	contestID, err := strconv.Atoi(unparsedContestID)

	if err != nil {
		res := response{
			Status:  "FAILED",
			Comment: "contestId: Field should contain long integer value",
		}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			logrus.WithError(err).Error()
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}

		return
	}

	ratingChanges := h.DB.GetRatingChanges(contestID)
	result, err := json.Marshal(ratingChanges)

	if err != nil {
		logrus.WithError(err).Error()
		w.WriteHeader(http.StatusInternalServerError)
	}

	res := response{
		Status: "OK",
		Result: result,
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		logrus.WithError(err).Error()
		w.WriteHeader(http.StatusInternalServerError)
	}

	return
}
