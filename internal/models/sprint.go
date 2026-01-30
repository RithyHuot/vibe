package models

import "time"

// Sprint represents a project sprint/iteration
type Sprint struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// IsActive checks if the sprint is currently active
func (s *Sprint) IsActive() bool {
	now := time.Now()
	return now.After(s.StartDate) && now.Before(s.EndDate)
}

// Folder represents a ClickUp folder (used for sprint detection)
type Folder struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
