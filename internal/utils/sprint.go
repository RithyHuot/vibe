package utils

import (
	"fmt"
	"regexp"
	"time"

	"github.com/rithyhuot/vibe/internal/models"
)

// SprintMatch represents a sprint with parsed dates
type SprintMatch struct {
	Folder    *models.Folder
	StartDate time.Time
	EndDate   time.Time
}

// FindCurrentSprintByDate finds the current sprint list by parsing date ranges from folder names
// Expected formats:
//   - "Sprint 5 (1/19 - 2/1)"
//   - "97 Commerce 1 (1/5 - 1/18)"
//   - Any folder with "(MM/DD - MM/DD)" pattern
func FindCurrentSprintByDate(folders []*models.Folder, patterns []string) *models.Folder {
	// Check cache first
	cacheKey := fmt.Sprintf("sprint:%v", patterns)
	if cached, found := GetSprintCache().Get(cacheKey); found {
		if folder, ok := cached.(*models.Folder); ok {
			// Verify the cached folder still exists in the list
			for _, f := range folders {
				if f.ID == folder.ID {
					return folder
				}
			}
		}
	}

	result := findCurrentSprintByDateUncached(folders, patterns)

	// Cache the result
	if result != nil {
		GetSprintCache().Set(cacheKey, result)
	}

	return result
}

//nolint:gocyclo // Sprint date range parsing requires multiple conditions
func findCurrentSprintByDateUncached(folders []*models.Folder, patterns []string) *models.Folder {
	today := time.Now()
	currentYear := today.Year()

	var matches []SprintMatch

	for _, folder := range folders {
		// If patterns provided, check if folder matches any pattern
		if len(patterns) > 0 {
			matchesPattern := false
			for _, pattern := range patterns {
				if matched, _ := regexp.MatchString(pattern, folder.Name); matched {
					matchesPattern = true
					break
				}
			}
			if !matchesPattern {
				continue
			}
		}

		// Extract date range: (MM/DD - MM/DD)
		dateRegex := regexp.MustCompile(`\((\d{1,2}/\d{1,2})\s*-\s*(\d{1,2}/\d{1,2})\)`)
		dateMatch := dateRegex.FindStringSubmatch(folder.Name)
		if dateMatch == nil {
			continue
		}

		startStr := dateMatch[1]
		endStr := dateMatch[2]

		// Parse dates
		startDate, err := parseSprintDate(startStr, currentYear)
		if err != nil {
			continue
		}

		endDate, err := parseSprintDate(endStr, currentYear)
		if err != nil {
			continue
		}

		// Set to start of day for start date and end of day for end date
		startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
		endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, endDate.Location())

		// Handle year boundary (e.g., Dec 15 - Jan 5)
		if endDate.Before(startDate) {
			// Sprint spans year boundary
			if today.Month() <= time.Month(endDate.Month()) {
				// We're in the new year part (Jan-Jun), so start was last year
				startDate = time.Date(currentYear-1, startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
			} else {
				// We're in the old year part (Jul-Dec), so end is next year
				endDate = time.Date(currentYear+1, endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, endDate.Location())
			}
		} else {
			// Dates don't cross year boundary, but we need to handle year transitions
			startMonth := int(startDate.Month())
			todayMonth := int(today.Month())

			// If sprint is in late year (Oct-Dec) and we're in early year (Jan-Mar),
			// the sprint was probably last year
			if startMonth >= 10 && todayMonth <= 3 {
				startDate = time.Date(currentYear-1, startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
				endDate = time.Date(currentYear-1, endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, endDate.Location())
			} else if startMonth <= 3 && todayMonth >= 10 {
				// If sprint is in early year (Jan-Mar) and we're in late year (Oct-Dec),
				// the sprint is probably next year
				startDate = time.Date(currentYear+1, startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
				endDate = time.Date(currentYear+1, endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, endDate.Location())
			}
		}

		matches = append(matches, SprintMatch{
			Folder:    folder,
			StartDate: startDate,
			EndDate:   endDate,
		})

		// Check if today falls within this range
		if (today.After(startDate) || today.Equal(startDate)) && (today.Before(endDate) || today.Equal(endDate)) {
			return folder
		}
	}

	// If no exact match, find the closest upcoming or most recent sprint
	if len(matches) > 0 {
		// Sort by start date
		for i := 0; i < len(matches)-1; i++ {
			for j := i + 1; j < len(matches); j++ {
				if matches[i].StartDate.After(matches[j].StartDate) {
					matches[i], matches[j] = matches[j], matches[i]
				}
			}
		}

		// First, try to find an upcoming sprint that starts soon (within 7 days)
		for _, match := range matches {
			if match.StartDate.After(today) {
				daysUntil := match.StartDate.Sub(today).Hours() / 24
				if daysUntil <= 7 {
					return match.Folder
				}
			}
		}

		// Otherwise, find the most recent sprint that has ended
		var lastPastSprint *SprintMatch
		for i := range matches {
			if matches[i].EndDate.Before(today) {
				lastPastSprint = &matches[i]
			}
		}
		if lastPastSprint != nil {
			return lastPastSprint.Folder
		}

		// If no past sprints, return the next upcoming one
		for _, match := range matches {
			if match.StartDate.After(today) {
				return match.Folder
			}
		}
	}

	return nil
}

// parseSprintDate parses a date string in M/D format with a given year
func parseSprintDate(dateStr string, year int) (time.Time, error) {
	dateRegex := regexp.MustCompile(`(\d{1,2})/(\d{1,2})`)
	matches := dateRegex.FindStringSubmatch(dateStr)
	if matches == nil {
		return time.Time{}, fmt.Errorf("invalid date format: %s", dateStr)
	}

	var month, day int
	_, err := parseInts(matches[1], matches[2], &month, &day)
	if err != nil {
		return time.Time{}, err
	}

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local), nil
}

// parseInts is a helper to parse multiple integer strings
func parseInts(strs ...interface{}) ([]int, error) {
	results := make([]int, 0, len(strs)/2)
	for i := 0; i < len(strs); i += 2 {
		str := strs[i].(string)
		ptr := strs[i+1].(*int)
		var val int
		_, err := fmt.Sscanf(str, "%d", &val)
		if err != nil {
			return nil, err
		}
		*ptr = val
		results = append(results, val)
	}
	return results, nil
}
