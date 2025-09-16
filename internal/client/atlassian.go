package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"atlassian-org-tool/internal/stats"
	"atlassian-org-tool/internal/types"
)

// AtlassianClient handles API requests to Atlassian
type AtlassianClient struct {
	config     types.Config
	httpClient *http.Client
}

// NewAtlassianClient creates a new Atlassian API client
func NewAtlassianClient(config types.Config) *AtlassianClient {
	return &AtlassianClient{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// makeRequest performs an HTTP request with retry logic and proper authentication
func (c *AtlassianClient) makeRequest(method, url string, bodyData []byte, useOAuth bool, tracker *stats.Tracker) ([]byte, error) {
	tracker.IncrementAPICallsTotal()
	var lastErr error
	maxRetries := 5 // Increased retry budget
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			tracker.IncrementRetryAttempts()
			// Exponential backoff with jitter
			delay := time.Duration(attempt*attempt) * time.Second
			time.Sleep(delay + time.Duration(attempt*100)*time.Millisecond)
		}

		var bodyReader io.Reader
		if bodyData != nil {
			bodyReader = bytes.NewReader(bodyData)
		}
		req, err := http.NewRequest(method, url, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		// Set authentication headers
		if useOAuth {
			if c.config.AccessToken == "" {
				return nil, fmt.Errorf("OAuth access token required for admin API")
			}
			req.Header.Set("Authorization", "Bearer "+c.config.AccessToken)
		} else {
			if c.config.Username == "" || c.config.APIToken == "" {
				return nil, fmt.Errorf("username and API token required for teams API")
			}
			req.SetBasicAuth(c.config.Username, c.config.APIToken)
		}

		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("making request: %w", err)
			continue
		}

		// Always read and close the response body to prevent connection leaks
		respBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()

		if readErr != nil {
			lastErr = fmt.Errorf("reading response: %w", readErr)
			continue
		}

		// Handle rate limiting with retry-after
		if resp.StatusCode == 429 {
			fmt.Fprintf(os.Stderr, "Rate limited (429), attempt %d/%d\n", attempt+1, maxRetries)
			retryAfter := resp.Header.Get("Retry-After")
			if retryAfter != "" {
				// Try parsing as seconds first
				if seconds, err := strconv.Atoi(retryAfter); err == nil {
					delay := time.Duration(seconds) * time.Second
					fmt.Fprintf(os.Stderr, "Waiting %v before retry\n", delay)
					time.Sleep(delay)
					continue
				}
				// Try parsing as HTTP date
				if retryTime, err := time.Parse(time.RFC1123, retryAfter); err == nil {
					delay := time.Until(retryTime)
					if delay > 0 && delay < 5*time.Minute { // Reasonable upper bound
						fmt.Fprintf(os.Stderr, "Waiting %v before retry\n", delay)
						time.Sleep(delay)
						continue
					}
				}
			}
			// Default backoff for 429 if header parsing fails
			delay := time.Duration((attempt+1)*10) * time.Second
			fmt.Fprintf(os.Stderr, "Using default backoff %v\n", delay)
			time.Sleep(delay)
			lastErr = fmt.Errorf("rate limited (429)")
			continue
		}

		// Handle server errors (5xx) with retry
		if resp.StatusCode >= 500 {
			fmt.Fprintf(os.Stderr, "Server error (%d), attempt %d/%d\n", resp.StatusCode, attempt+1, maxRetries)
			lastErr = fmt.Errorf("server error: status %d", resp.StatusCode)
			continue
		}

		// Handle client errors (4xx) - don't retry these
		if resp.StatusCode >= 400 {
			tracker.IncrementAPICallsFailed()
			return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
		}

		// Success case
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			tracker.IncrementAPICallsSucceeded()
			return respBody, nil
		}

		// Unexpected status code
		lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	tracker.IncrementAPICallsFailed()
	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// GetOrganization retrieves organization information using admin API
func (c *AtlassianClient) GetOrganization(tracker *stats.Tracker) (*types.Organization, error) {
	url := fmt.Sprintf("%s/admin/v1/orgs/%s", c.config.AdminBaseURL, c.config.OrgID)

	body, err := c.makeRequest("GET", url, nil, true, tracker) // Use OAuth
	if err != nil {
		return nil, fmt.Errorf("getting organization: %w", err)
	}

	var org types.Organization
	if err := json.Unmarshal(body, &org); err != nil {
		return nil, fmt.Errorf("unmarshaling organization: %w", err)
	}

	tracker.IncrementOrganizations()
	return &org, nil
}

// TeamsResponse represents the response from the teams API
type TeamsResponse struct {
	Data []types.Team `json:"data"`
	Meta struct {
		Total     int `json:"total"`
		TotalPage int `json:"totalPage"`
	} `json:"meta"`
}

// GetTeams retrieves all teams for the organization using teams API with pagination
func (c *AtlassianClient) GetTeams(tracker *stats.Tracker) ([]types.Team, error) {
	baseURL := fmt.Sprintf("https://api.atlassian.com/ex/teams/%s/teams", c.config.OrgID)

	var allTeams []types.Team
	page := 1
	limit := 50 // API default/max

	for {
		// Build URL with pagination parameters
		reqURL, err := url.Parse(baseURL)
		if err != nil {
			return nil, fmt.Errorf("parsing base URL: %w", err)
		}

		params := url.Values{}
		params.Set("page", strconv.Itoa(page))
		params.Set("limit", strconv.Itoa(limit))
		reqURL.RawQuery = params.Encode()

		body, err := c.makeRequest("GET", reqURL.String(), nil, false, tracker) // Use basic auth
		if err != nil {
			return nil, fmt.Errorf("getting teams (page %d): %w", page, err)
		}

		var teamsResp TeamsResponse
		if err := json.Unmarshal(body, &teamsResp); err != nil {
			return nil, fmt.Errorf("unmarshaling teams response (page %d): %w", page, err)
		}

		allTeams = append(allTeams, teamsResp.Data...)
		tracker.IncrementTeams() // Add the number of teams found in this page

		// Check if we've got all teams
		if len(allTeams) >= teamsResp.Meta.Total || len(teamsResp.Data) == 0 {
			break
		}

		page++
	}

	// Set Name field for internal use
	for i := range allTeams {
		allTeams[i].Name = allTeams[i].DisplayName
	}

	return allTeams, nil
}

// MembersRequest represents the request body for getting team members
type MembersRequest struct {
	TeamIds []string `json:"teamIds"`
}

// MembersResponse represents the response from the team members API
type MembersResponse struct {
	Data []struct {
		TeamID  string         `json:"teamId"`
		Members []types.Member `json:"members"`
	} `json:"data"`
	Meta struct {
		Next string `json:"next,omitempty"`
	} `json:"meta"`
}

// GetTeamMembers gets members for a specific team using the teams API with pagination
func (c *AtlassianClient) GetTeamMembers(teamID string, tracker *stats.Tracker) ([]types.Member, error) {
	baseURL := fmt.Sprintf("https://api.atlassian.com/ex/teams/%s/members", c.config.OrgID)

	var allMembers []types.Member
	var cursor string

	for {
		// Prepare request body
		reqBody := MembersRequest{
			TeamIds: []string{teamID},
		}

		bodyData, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("marshaling members request: %w", err)
		}

		// Build URL with cursor if available
		reqURL := baseURL
		if cursor != "" {
			reqURL += "?cursor=" + url.QueryEscape(cursor)
		}

		body, err := c.makeRequest("POST", reqURL, bodyData, false, tracker) // Use basic auth
		if err != nil {
			return nil, fmt.Errorf("getting team members for team %s: %w", teamID, err)
		}

		var membersResp MembersResponse
		if err := json.Unmarshal(body, &membersResp); err != nil {
			return nil, fmt.Errorf("unmarshaling members response for team %s: %w", teamID, err)
		}

		// Find our team's data in the response
		for _, teamData := range membersResp.Data {
			if teamData.TeamID == teamID {
				allMembers = append(allMembers, teamData.Members...)
				break
			}
		}

		// Check for pagination
		if membersResp.Meta.Next == "" {
			break
		}
		cursor = strings.TrimSpace(membersResp.Meta.Next)
	}

	tracker.IncrementMembers(len(allMembers))
	return allMembers, nil
}
