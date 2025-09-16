package types

// Config holds the configuration for the Atlassian API client
type Config struct {
	AdminBaseURL string
	AccessToken  string // OAuth 2.0 access token for admin APIs
	Username     string // For basic auth with teams API
	APIToken     string // For basic auth with teams API
	OrgID        string
}

// Organization represents an Atlassian organization
type Organization struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	DisplayName string                 `json:"displayName"`
	Type        string                 `json:"type"`
	Links       map[string]interface{} `json:"links,omitempty"`
}

// Team represents an Atlassian team (matching API response format)
type Team struct {
	ID             string                 `json:"teamId"`
	DisplayName    string                 `json:"displayName"`
	Description    string                 `json:"description,omitempty"`
	Type           string                 `json:"teamType"`
	OrganizationID string                 `json:"organizationId"`
	CreatorID      string                 `json:"creatorId,omitempty"`
	Members        []Member               `json:"members,omitempty"`
	Links          map[string]interface{} `json:"links,omitempty"`
	// Name is an alias for DisplayName for internal use
	Name string `json:"-"`
}

// Member represents a team member
type Member struct {
	AccountID   string `json:"accountId"`
	Email       string `json:"email,omitempty"`
	DisplayName string `json:"displayName"`
	AccountType string `json:"accountType"`
}

// NDJSONOutput represents the output format for streaming
type NDJSONOutput struct {
	Type      string      `json:"type"`
	Timestamp string      `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// Statistics represents processing statistics
type Statistics struct {
	StartTime          string `json:"startTime"`
	OrganizationsFound int    `json:"organizationsFound"`
	TeamsFound         int    `json:"teamsFound"`
	TeamsProcessed     int    `json:"teamsProcessed"`
	MembersFound       int    `json:"membersFound"`
	APICallsTotal      int    `json:"apiCallsTotal"`
	APICallsSucceeded  int    `json:"apiCallsSucceeded"`
	APICallsFailed     int    `json:"apiCallsFailed"`
	RetryAttempts      int    `json:"retryAttempts"`
	ProcessingTime     string `json:"processingTime"`
	Status             string `json:"status"`
	LastUpdated        string `json:"lastUpdated"`
}

// JSONSchema represents the JSON schema for the NDJSON output
type JSONSchema struct {
	Schema               string                 `json:"$schema"`
	Title                string                 `json:"title"`
	Description          string                 `json:"description"`
	Type                 string                 `json:"type"`
	Properties           map[string]interface{} `json:"properties"`
	Required             []string               `json:"required"`
	AdditionalProperties bool                   `json:"additionalProperties"`
}
