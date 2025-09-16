package stats

import (
	"testing"
	"time"
)

func TestNewTracker(t *testing.T) {
	tracker := NewTracker()

	if tracker == nil {
		t.Fatal("NewTracker returned nil")
	}

	stats := tracker.GetStats()
	if stats.Status != "processing" {
		t.Errorf("Expected initial status 'processing', got '%s'", stats.Status)
	}
	if stats.OrganizationsFound != 0 {
		t.Errorf("Expected OrganizationsFound to be 0, got %d", stats.OrganizationsFound)
	}
	if stats.TeamsFound != 0 {
		t.Errorf("Expected TeamsFound to be 0, got %d", stats.TeamsFound)
	}
}

func TestTrackerIncrements(t *testing.T) {
	tracker := NewTracker()

	// Test organization increment
	tracker.IncrementOrganizations()
	stats := tracker.GetStats()
	if stats.OrganizationsFound != 1 {
		t.Errorf("Expected OrganizationsFound to be 1, got %d", stats.OrganizationsFound)
	}

	// Test teams increment
	tracker.IncrementTeams()
	tracker.IncrementTeams()
	stats = tracker.GetStats()
	if stats.TeamsFound != 2 {
		t.Errorf("Expected TeamsFound to be 2, got %d", stats.TeamsFound)
	}

	// Test teams processed increment
	tracker.IncrementTeamsProcessed()
	stats = tracker.GetStats()
	if stats.TeamsProcessed != 1 {
		t.Errorf("Expected TeamsProcessed to be 1, got %d", stats.TeamsProcessed)
	}

	// Test members increment
	tracker.IncrementMembers(5)
	stats = tracker.GetStats()
	if stats.MembersFound != 5 {
		t.Errorf("Expected MembersFound to be 5, got %d", stats.MembersFound)
	}

	// Test API calls
	tracker.IncrementAPICallsTotal()
	tracker.IncrementAPICallsSucceeded()
	tracker.IncrementAPICallsTotal()
	tracker.IncrementAPICallsFailed()
	stats = tracker.GetStats()
	if stats.APICallsTotal != 2 {
		t.Errorf("Expected APICallsTotal to be 2, got %d", stats.APICallsTotal)
	}
	if stats.APICallsSucceeded != 1 {
		t.Errorf("Expected APICallsSucceeded to be 1, got %d", stats.APICallsSucceeded)
	}
	if stats.APICallsFailed != 1 {
		t.Errorf("Expected APICallsFailed to be 1, got %d", stats.APICallsFailed)
	}

	// Test retry attempts
	tracker.IncrementRetryAttempts()
	tracker.IncrementRetryAttempts()
	stats = tracker.GetStats()
	if stats.RetryAttempts != 2 {
		t.Errorf("Expected RetryAttempts to be 2, got %d", stats.RetryAttempts)
	}
}

func TestTrackerStatusUpdate(t *testing.T) {
	tracker := NewTracker()

	// Test status change
	tracker.SetStatus("completed")
	stats := tracker.GetStats()
	if stats.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", stats.Status)
	}

	tracker.SetStatus("failed")
	stats = tracker.GetStats()
	if stats.Status != "failed" {
		t.Errorf("Expected status 'failed', got '%s'", stats.Status)
	}
}

func TestTrackerUpdateStats(t *testing.T) {
	tracker := NewTracker()

	// Wait enough time to ensure there's a measureable difference
	time.Sleep(1 * time.Second)

	// Update stats and check processing time
	updatedStats := tracker.UpdateStats()
	if updatedStats.ProcessingTime == "0s" {
		t.Error("Expected ProcessingTime to be non-zero after UpdateStats")
	}

	// Check that LastUpdated is different from StartTime
	if updatedStats.LastUpdated == updatedStats.StartTime {
		t.Error("Expected LastUpdated to be different from StartTime after UpdateStats")
	}
}

func TestTrackerStatsCopy(t *testing.T) {
	tracker := NewTracker()

	// Modify the tracker
	tracker.IncrementOrganizations()
	tracker.SetStatus("completed")

	// Get stats and verify they're a copy
	stats1 := tracker.GetStats()
	stats2 := tracker.GetStats()

	// Verify values are same
	if stats1.OrganizationsFound != stats2.OrganizationsFound {
		t.Error("Stats should be consistent between calls")
	}
	if stats1.Status != stats2.Status {
		t.Error("Status should be consistent between calls")
	}

	// Modify tracker again
	tracker.IncrementOrganizations()
	stats3 := tracker.GetStats()

	// Original stats should be unchanged (confirming they're copies)
	if stats1.OrganizationsFound != 1 {
		t.Error("Original stats should not be modified")
	}
	if stats3.OrganizationsFound != 2 {
		t.Error("New stats should reflect changes")
	}
}
