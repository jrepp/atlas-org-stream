package stats

import (
	"time"

	"atlassian-org-tool/internal/types"
)

// Tracker holds processing statistics
type Tracker struct {
	stats     types.Statistics
	startTime time.Time
}

// NewTracker creates a new statistics tracker
func NewTracker() *Tracker {
	now := time.Now()
	return &Tracker{
		stats: types.Statistics{
			StartTime:          now.Format(time.RFC3339),
			OrganizationsFound: 0,
			TeamsFound:         0,
			TeamsProcessed:     0,
			MembersFound:       0,
			APICallsTotal:      0,
			APICallsSucceeded:  0,
			APICallsFailed:     0,
			RetryAttempts:      0,
			ProcessingTime:     "0s",
			Status:             "processing",
			LastUpdated:        now.Format(time.RFC3339),
		},
		startTime: now,
	}
}

// UpdateStats updates the statistics and returns the current state
func (t *Tracker) UpdateStats() types.Statistics {
	now := time.Now()
	t.stats.ProcessingTime = now.Sub(t.startTime).String()
	t.stats.LastUpdated = now.Format(time.RFC3339)
	return t.stats
}

// IncrementOrganizations increments the organizations found counter
func (t *Tracker) IncrementOrganizations() {
	t.stats.OrganizationsFound++
}

// IncrementTeams increments the teams found counter
func (t *Tracker) IncrementTeams() {
	t.stats.TeamsFound++
}

// IncrementTeamsProcessed increments the teams processed counter
func (t *Tracker) IncrementTeamsProcessed() {
	t.stats.TeamsProcessed++
}

// IncrementMembers increments the members found counter by the given amount
func (t *Tracker) IncrementMembers(count int) {
	t.stats.MembersFound += count
}

// IncrementAPICallsTotal increments the total API calls counter
func (t *Tracker) IncrementAPICallsTotal() {
	t.stats.APICallsTotal++
}

// IncrementAPICallsSucceeded increments the successful API calls counter
func (t *Tracker) IncrementAPICallsSucceeded() {
	t.stats.APICallsSucceeded++
}

// IncrementAPICallsFailed increments the failed API calls counter
func (t *Tracker) IncrementAPICallsFailed() {
	t.stats.APICallsFailed++
}

// IncrementRetryAttempts increments the retry attempts counter
func (t *Tracker) IncrementRetryAttempts() {
	t.stats.RetryAttempts++
}

// SetStatus sets the processing status
func (t *Tracker) SetStatus(status string) {
	t.stats.Status = status
}

// GetStats returns a copy of the current statistics
func (t *Tracker) GetStats() types.Statistics {
	return t.stats
}
