package domain

import (
	"fmt"
	"math/rand"
	"time"
)

var TimeRanges = []string{
	"10:00 -- 12:00",
	"12:00 -- 14:00",
	"14:00 -- 16:00",
	"16:00 -- 18:00",
	"18:00 -- 20:00",
	"20:00 -- 22:00",
}

func Timef(t time.Time) string {
	loc, _ := time.LoadLocation("Europe/Samara")
	return t.In(loc).Format("02.01 в 15:04")
}

func BinaryToSet(binary string) map[string]bool {
	selected := make(map[string]bool)
	for i, bit := range binary {
		if bit == '1' && i < len(TimeRanges) {
			selected[TimeRanges[i]] = true
		}
	}
	return selected
}

func SetToBinary(selected map[string]bool) string {
	result := make([]byte, len(TimeRanges))
	for i, tr := range TimeRanges {
		if selected[tr] {
			result[i] = '1'
		} else {
			result[i] = '0'
		}
	}
	return string(result)
}

func PickRandomTime(timeIntersection string) string {
	var indices []int
	for i, bit := range timeIntersection {
		if bit == '1' {
			indices = append(indices, i)
		}
	}

	if len(indices) == 0 {
		return "12:00"
	}

	index := indices[rand.Intn(len(indices))]
	tr := TimeRanges[index]

	begin := tr[:2]
	minutes := rand.Intn(12) * 5
	// minutes := rand.Intn(6) * 10

	return fmt.Sprintf("%s:%02d", begin, minutes)
}

func HasTimeOverlap(timeRange string) bool {
	for _, ch := range timeRange {
		if ch == '1' {
			return true
		}
	}
	return false
}

func MergeSelectedRanges(selected map[string]bool) []string {
	var merged []string
	start := -1
	for i, tr := range TimeRanges {
		if selected[tr] {
			if start == -1 {
				start = i
			}
		} else {
			if start != -1 {
				merged = append(merged, TimeRanges[start][:5]+" -- "+TimeRanges[i-1][len(TimeRanges[i-1])-5:])
				start = -1
			}
		}
	}
	if start != -1 {
		last := TimeRanges[len(TimeRanges)-1]
		merged = append(merged, TimeRanges[start][:5]+" -- "+last[len(last)-5:])
	}
	return merged
}

func CalculateTimeIntersection(a, b string) string {
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
