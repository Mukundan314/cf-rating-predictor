package calculator

import (
	"math"
	"sort"

	"github.com/cheran-senthil/cf-rating-predictor/fft"
	"github.com/mukundan314/go-codeforces"
)

type contestant struct {
	party      codeforces.Party
	rank       float64
	points     float64
	penalty    int
	rating     int
	needRating int
	delta      int
	seed       float64
}

const max = 8192 // The maximum rating difference

var eloWinProbability = make([]float64, 2*max)
var eloWinProbabilityFFT = make([]complex128, 2*max)

func init() {
	eloWinProbability[0] = getEloWinProbability(0, 0)
	eloWinProbabilityFFT[0] = complex(eloWinProbability[0], 0)

	for i := 1; i <= max; i++ {
		eloWinProbability[i] = getEloWinProbability(0, float64(i))
		eloWinProbabilityFFT[i] = complex(eloWinProbability[i], 0)

		eloWinProbability[2*max-i] = 1 - eloWinProbability[i]
		eloWinProbabilityFFT[2*max-i] = 1 - eloWinProbabilityFFT[i]
	}

	eloWinProbabilityFFT = fft.FFT(eloWinProbabilityFFT, false)
}

func intMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func intMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func CalculateRatingChanges(previousRatings map[string]int, standingsRows []codeforces.RanklistRow) map[string]int {
	contestants := make([]*contestant, 0, len(standingsRows)-1)

	for _, standingsRow := range standingsRows {
		var rating int

		if len(standingsRow.Party.Members) > 1 {
			ratings := make([]int, len(standingsRow.Party.Members))
			for j, member := range standingsRow.Party.Members {
				ratings[j] = previousRatings[member.Handle]
			}

			rating = ComposeRatingsByTeamMemberRatings(ratings)
		} else {
			rating = previousRatings[standingsRow.Party.Members[0].Handle]
		}

		contestants = append(contestants, &contestant{
			party:   standingsRow.Party,
			rank:    float64(standingsRow.Rank),
			points:  standingsRow.Points,
			penalty: standingsRow.Penalty,
			rating:  rating,
		})
	}

	process(contestants)

	ratingChanges := make(map[string]int)
	for _, contestant := range contestants {
		for _, member := range contestant.party.Members {
			ratingChanges[member.Handle] = contestant.delta
		}
	}

	return ratingChanges
}

func getEloWinProbability(ra, rb float64) float64 {
	return 1 / (1 + math.Pow(10, (rb-ra)/400))
}

func ComposeRatingsByTeamMemberRatings(ratings []int) int {
	left := 100.0
	right := 4000.0

	for tt := 0; tt < 20; tt++ {
		r := (left + right) / 2

		rWinsProbability := 1.0
		for _, rating := range ratings {
			rWinsProbability *= getEloWinProbability(r, float64(rating))
		}

		rating := math.Log10(1/(rWinsProbability)-1)*400 + r

		if rating > r {
			left = r
		} else {
			right = r
		}
	}

	return int(math.Round((left + right) / 2))
}

func calculateSeeds(contestants []*contestant) []float64 {
	count := make([]complex128, 2*max)
	for _, contestant := range contestants {
		count[((2*max)+contestant.rating)%(2*max)]++
	}

	countFFT := fft.FFT(count, false)

	seedsFFT := make([]complex128, 2*max)
	for i, v := range countFFT {
		seedsFFT[i] = v * eloWinProbabilityFFT[i]
	}

	seedsCmplx := fft.FFT(seedsFFT, true)
	seeds := make([]float64, 2*max)

	for i, v := range seedsCmplx {
		seeds[i] = 1 + real(v)
	}

	return seeds
}

func reassignRanks(contestants []*contestant) {
	sortByPointsDesc(contestants)

	points := contestants[len(contestants)-1].points
	penalty := contestants[len(contestants)-1].penalty
	rank := float64(len(contestants))

	for i := len(contestants) - 1; i >= 0; i-- {
		if contestants[i].points != points || contestants[i].penalty != penalty {
			rank = float64(i) + 1.0
			points = contestants[i].points
			penalty = contestants[i].penalty
		}
		contestants[i].rank = rank
	}
}

func sortByPointsDesc(contestants []*contestant) {
	sort.Slice(contestants, func(i, j int) bool {
		if contestants[i].points == contestants[j].points {
			return contestants[i].penalty < contestants[j].penalty
		}

		return contestants[i].points > contestants[j].points
	})
}

func process(contestants []*contestant) {
	if len(contestants) == 0 {
		return
	}

	reassignRanks(contestants)

	seeds := calculateSeeds(contestants)

	for _, contestant := range contestants {
		contestant.seed = seeds[(2*max+contestant.rating)%(2*max)] - eloWinProbability[0]
		midRank := math.Sqrt(contestant.rank * contestant.seed)

		left, right := 1, 8000
		for right-left > 1 {
			mid := (left + right) / 2
			if seeds[mid]-eloWinProbability[(2*max+(mid-contestant.rating))%(2*max)] < midRank {
				right = mid
			} else {
				left = mid
			}
		}
		contestant.needRating = left

		contestant.delta = (contestant.needRating - contestant.rating) / 2
	}

	sortByRatingDesc(contestants)

	{
		sum := 0
		for _, contestant := range contestants {
			sum += contestant.delta
		}

		inc := -sum/len(contestants) - 1
		for _, contestant := range contestants {
			contestant.delta += inc
		}
	}

	{
		sum := 0
		zeroSumCount := int(math.Min(4*math.Round(math.Sqrt(float64(len(contestants)))), float64(len(contestants))))
		for i := 0; i < zeroSumCount; i++ {
			sum += contestants[i].delta
		}
		inc := intMin(intMax(-sum/zeroSumCount, -10), 0)
		for _, contestant := range contestants {
			contestant.delta += inc
		}
	}
}

func sortByRatingDesc(contestants []*contestant) {
	sort.Slice(contestants, func(i, j int) bool {
		return contestants[i].rating > contestants[j].rating
	})
}
