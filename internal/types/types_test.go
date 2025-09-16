package types

import (
	"encoding/json"
	"testing"
)

func TestOrganizationJSONSerialization(t *testing.T) {
	org := Organization{
		ID:          "test-org-id",
		Name:        "Test Organization",
		DisplayName: "Test Org Display",
		Type:        "organization",
		Links:       map[string]interface{}{"self": "https://example.com/org"},
	}

	// Test JSON marshaling
	data, err := json.Marshal(org)
	if err != nil {
		t.Fatalf("Failed to marshal Organization: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaledOrg Organization
	err = json.Unmarshal(data, &unmarshaledOrg)
	if err != nil {
		t.Fatalf("Failed to unmarshal Organization: %v", err)
	}

	// Verify fields
	if unmarshaledOrg.ID != org.ID {
		t.Errorf("Expected ID %s, got %s", org.ID, unmarshaledOrg.ID)
	}
	if unmarshaledOrg.Name != org.Name {
		t.Errorf("Expected Name %s, got %s", org.Name, unmarshaledOrg.Name)
	}
	if unmarshaledOrg.DisplayName != org.DisplayName {
		t.Errorf("Expected DisplayName %s, got %s", org.DisplayName, unmarshaledOrg.DisplayName)
	}
}

func TestTeamJSONSerialization(t *testing.T) {
	team := Team{
		ID:             "test-team-id",
		DisplayName:    "Test Team",
		Description:    "A test team",
		Type:           "team",
		OrganizationID: "test-org-id",
		CreatorID:      "test-creator-id",
		Members: []Member{
			{
				AccountID:   "member1",
				Email:       "member1@example.com",
				DisplayName: "Member One",
				AccountType: "atlassian",
			},
		},
		Name: "Test Team Internal",
	}

	// Test JSON marshaling
	data, err := json.Marshal(team)
	if err != nil {
		t.Fatalf("Failed to marshal Team: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaledTeam Team
	err = json.Unmarshal(data, &unmarshaledTeam)
	if err != nil {
		t.Fatalf("Failed to unmarshal Team: %v", err)
	}

	// Verify fields
	if unmarshaledTeam.ID != team.ID {
		t.Errorf("Expected ID %s, got %s", team.ID, unmarshaledTeam.ID)
	}
	if unmarshaledTeam.DisplayName != team.DisplayName {
		t.Errorf("Expected DisplayName %s, got %s", team.DisplayName, unmarshaledTeam.DisplayName)
	}
	if len(unmarshaledTeam.Members) != len(team.Members) {
		t.Errorf("Expected %d members, got %d", len(team.Members), len(unmarshaledTeam.Members))
	}
	// Note: Name field should not be marshaled due to json:"-" tag
	if unmarshaledTeam.Name != "" {
		t.Errorf("Expected Name to be empty after unmarshaling, got %s", unmarshaledTeam.Name)
	}
}

func TestStatisticsJSONSerialization(t *testing.T) {
	stats := Statistics{
		StartTime:          "2024-01-01T00:00:00Z",
		OrganizationsFound: 1,
		TeamsFound:         5,
		TeamsProcessed:     3,
		MembersFound:       15,
		APICallsTotal:      10,
		APICallsSucceeded:  8,
		APICallsFailed:     2,
		RetryAttempts:      1,
		ProcessingTime:     "5m30s",
		Status:             "processing",
		LastUpdated:        "2024-01-01T00:05:30Z",
	}

	// Test JSON marshaling
	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("Failed to marshal Statistics: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaledStats Statistics
	err = json.Unmarshal(data, &unmarshaledStats)
	if err != nil {
		t.Fatalf("Failed to unmarshal Statistics: %v", err)
	}

	// Verify fields
	if unmarshaledStats.StartTime != stats.StartTime {
		t.Errorf("Expected StartTime %s, got %s", stats.StartTime, unmarshaledStats.StartTime)
	}
	if unmarshaledStats.OrganizationsFound != stats.OrganizationsFound {
		t.Errorf("Expected OrganizationsFound %d, got %d", stats.OrganizationsFound, unmarshaledStats.OrganizationsFound)
	}
	if unmarshaledStats.Status != stats.Status {
		t.Errorf("Expected Status %s, got %s", stats.Status, unmarshaledStats.Status)
	}
}

func TestNDJSONOutputStructure(t *testing.T) {
	// Test with organization data
	org := Organization{ID: "test", Name: "Test", DisplayName: "Test Org", Type: "org"}
	output := NDJSONOutput{
		Type:      "organization",
		Timestamp: "2024-01-01T00:00:00Z",
		Data:      org,
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal NDJSONOutput: %v", err)
	}

	var unmarshaledOutput NDJSONOutput
	err = json.Unmarshal(data, &unmarshaledOutput)
	if err != nil {
		t.Fatalf("Failed to unmarshal NDJSONOutput: %v", err)
	}

	if unmarshaledOutput.Type != output.Type {
		t.Errorf("Expected Type %s, got %s", output.Type, unmarshaledOutput.Type)
	}
	if unmarshaledOutput.Timestamp != output.Timestamp {
		t.Errorf("Expected Timestamp %s, got %s", output.Timestamp, unmarshaledOutput.Timestamp)
	}
}