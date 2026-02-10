package matcher

import (
	"fmt"
	"math"
	"regexp"

	"github.com/jus1d/kypidbot/internal/infrastructure/ollama"
)

type MatchUser struct {
	Index      int
	Username   string
	Sex        string
	About      string
	TimeRanges string
}

type MatchPair struct {
	I                int
	J                int
	Score            float64
	TimeIntersection string
}

type FullMatch struct {
	I     int
	J     int
	Score float64
}

// Match matches all users based on product logic: different genders, hungarian algorithm etc.
func Match(users []MatchUser, ollama *ollama.Client) ([]MatchPair, []FullMatch, error) {
	if len(users) < 2 {
		return nil, nil, fmt.Errorf("need at least 2 users")
	}

	abouts := make([]string, len(users))
	for i, u := range users {
		abouts[i] = u.About
	}

	vectors, err := ollama.GetEmbeddings(abouts)
	if err != nil {
		return nil, nil, fmt.Errorf("get embeddings: %w", err)
	}

	simMatrix := make([][]float64, len(vectors))
	for i := range simMatrix {
		simMatrix[i] = make([]float64, len(vectors))
		for j := range simMatrix[i] {
			simMatrix[i][j] = cosineSimilarity(vectors[i], vectors[j])
		}
	}

	preferences := extractPreferences(users)

	used := make(map[int]bool)
	var pairs []MatchPair
	var fullMatches []FullMatch

	n := len(users)
	for i := 0; i < n; i++ {
		if used[i] {
			continue
		}
		for j := i + 1; j < n; j++ {
			if used[j] {
				continue
			}

			a, b := users[i], users[j]
			if a.Sex == b.Sex {
				continue
			}

			aWantsB := preferences[i] != nil && preferences[i][b.Username]
			bWantsA := preferences[j] != nil && preferences[j][a.Username]

			if aWantsB && bWantsA {
				pairTime := calculateTimeIntersection(a.TimeRanges, b.TimeRanges)
				score := simMatrix[i][j]

				if hasTimeOverlap(pairTime) {
					pairs = append(pairs, MatchPair{
						I:                i,
						J:                j,
						Score:            math.Round(score*1000) / 1000,
						TimeIntersection: pairTime,
					})
				} else {
					fullMatches = append(fullMatches, FullMatch{
						I:     i,
						J:     j,
						Score: math.Round(score*1000) / 1000,
					})
				}

				used[i] = true
				used[j] = true
				break
			}
		}
	}

	var males, females []int
	for i := 0; i < n; i++ {
		if used[i] {
			continue
		}
		if users[i].Sex == "male" {
			males = append(males, i)
		} else {
			females = append(females, i)
		}
	}

	size := len(males)
	if len(females) > size {
		size = len(females)
	}

	if size == 0 {
		return pairs, fullMatches, nil
	}

	scoreMatrix := make([][]float64, size)
	for i := range scoreMatrix {
		scoreMatrix[i] = make([]float64, size)
		for j := range scoreMatrix[i] {
			if i >= len(males) || j >= len(females) {
				scoreMatrix[i][j] = -1e9
				continue
			}

			mi, fj := males[i], females[j]
			pairTime := calculateTimeIntersection(users[mi].TimeRanges, users[fj].TimeRanges)

			if !hasTimeOverlap(pairTime) {
				scoreMatrix[i][j] = -1e9
				continue
			}

			score := simMatrix[mi][fj]

			aWantsB := preferences[mi] != nil && preferences[mi][users[fj].Username]
			bWantsA := preferences[fj] != nil && preferences[fj][users[mi].Username]
			if aWantsB || bWantsA {
				score += 0.3
			}

			scoreMatrix[i][j] = score
		}
	}

	assignment := hungarian(scoreMatrix)

	for i, j := range assignment {
		if i >= len(males) || j >= len(females) {
			continue
		}
		mi, fj := males[i], females[j]
		pairTime := calculateTimeIntersection(users[mi].TimeRanges, users[fj].TimeRanges)
		if !hasTimeOverlap(pairTime) {
			continue
		}

		score := simMatrix[mi][fj]
		aWantsB := preferences[mi] != nil && preferences[mi][users[fj].Username]
		bWantsA := preferences[fj] != nil && preferences[fj][users[mi].Username]
		if aWantsB || bWantsA {
			score += 0.3
		}

		pairs = append(pairs, MatchPair{
			I:                mi,
			J:                fj,
			Score:            math.Round(score*1000) / 1000,
			TimeIntersection: pairTime,
		})
	}

	return pairs, fullMatches, nil
}

type ScorePair struct {
	A     string
	B     string
	Score float64
}

// MatchByScore matches abouts into pairs purely by similarity score using Hungarian algorithm.
func MatchByScore(abouts []string, ollama *ollama.Client) ([]ScorePair, error) {
	if len(abouts) < 2 {
		return nil, fmt.Errorf("need at least 2 abouts")
	}

	vectors, err := ollama.GetEmbeddings(abouts)
	if err != nil {
		return nil, fmt.Errorf("get embeddings: %w", err)
	}

	n := len(abouts)
	size := n / 2
	if n%2 != 0 {
		size = (n + 1) / 2
	}

	scoreMatrix := make([][]float64, size)
	for i := range scoreMatrix {
		scoreMatrix[i] = make([]float64, size)
		for j := range scoreMatrix[i] {
			ai := i
			bj := size + j
			if ai >= n || bj >= n {
				scoreMatrix[i][j] = -1e9
				continue
			}
			scoreMatrix[i][j] = cosineSimilarity(vectors[ai], vectors[bj])
		}
	}

	assignment := hungarian(scoreMatrix)

	var pairs []ScorePair
	for i, j := range assignment {
		bj := size + j
		if i >= n || bj >= n {
			continue
		}
		score := cosineSimilarity(vectors[i], vectors[bj])
		pairs = append(pairs, ScorePair{
			A:     abouts[i],
			B:     abouts[bj],
			Score: math.Round(score*1000) / 1000,
		})
	}

	return pairs, nil
}

func extractPreferences(users []MatchUser) map[int]map[string]bool {
	preferences := make(map[int]map[string]bool)
	pattern := regexp.MustCompile(`@(\w+)`)

	for i, user := range users {
		matches := pattern.FindAllStringSubmatch(user.About, -1)
		if len(matches) > 0 {
			preferences[i] = make(map[string]bool)
			for _, match := range matches {
				preferences[i][match[1]] = true
			}
		}
	}

	return preferences
}

func calculateTimeIntersection(a, b string) string {
	if len(a) != 6 || len(b) != 6 {
		return "000000"
	}

	result := make([]byte, 6)
	for k := 0; k < 6; k++ {
		if a[k] == '1' && b[k] == '1' {
			result[k] = '1'
		} else {
			result[k] = '0'
		}
	}
	return string(result)
}

func hasTimeOverlap(timeRange string) bool {
	for _, ch := range timeRange {
		if ch == '1' {
			return true
		}
	}
	return false
}
