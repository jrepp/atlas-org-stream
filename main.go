package main

import (
        "bytes"
        "encoding/json"
        "flag"
        "fmt"
        "io"
        "log"
        "net/http"
        "net/url"
        "os"
        "strconv"
        "strings"
        "time"
)

// Config holds the configuration for the Atlassian API client
type Config struct {
        AdminBaseURL string
        AccessToken  string // OAuth 2.0 access token for admin APIs
        Username     string // For basic auth with teams API
        APIToken     string // For basic auth with teams API
        OrgID        string
}

// AtlassianClient handles API requests to Atlassian
type AtlassianClient struct {
        config     Config
        httpClient *http.Client
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
        Schema              string                 `json:"$schema"`
        Title               string                 `json:"title"`
        Description         string                 `json:"description"`
        Type                string                 `json:"type"`
        Properties          map[string]interface{} `json:"properties"`
        Required            []string               `json:"required"`
        AdditionalProperties bool                   `json:"additionalProperties"`
}

// StatisticsTracker holds processing statistics
type StatisticsTracker struct {
        stats     Statistics
        startTime time.Time
}

// NewStatisticsTracker creates a new statistics tracker
func NewStatisticsTracker() *StatisticsTracker {
        now := time.Now()
        return &StatisticsTracker{
                stats: Statistics{
                        StartTime:   now.UTC().Format(time.RFC3339),
                        Status:      "processing",
                        LastUpdated: now.UTC().Format(time.RFC3339),
                },
                startTime: now,
        }
}

// UpdateStats updates the statistics and returns a copy
func (st *StatisticsTracker) UpdateStats() Statistics {
        now := time.Now()
        st.stats.ProcessingTime = now.Sub(st.startTime).String()
        st.stats.LastUpdated = now.UTC().Format(time.RFC3339)
        return st.stats
}

// createJSONSchema creates the JSON schema for the NDJSON output
func createJSONSchema() JSONSchema {
        return JSONSchema{
                Schema:      "https://json-schema.org/draft/2020-12/schema",
                Title:       "Atlassian Organization and Team Data Stream",
                Description: "Schema for NDJSON stream containing Atlassian organization and team metadata",
                Type:        "object",
                Required:    []string{"type", "timestamp", "data"},
                AdditionalProperties: false,
                Properties: map[string]interface{}{
                        "type": map[string]interface{}{
                                "type": "string",
                                "enum": []string{"schema", "organization", "team", "statistics"},
                                "description": "Type of data in this line",
                        },
                        "timestamp": map[string]interface{}{
                                "type": "string",
                                "format": "date-time",
                                "description": "ISO 8601 timestamp when this data was generated",
                        },
                        "data": map[string]interface{}{
                                "oneOf": []map[string]interface{}{
                                        {
                                                "type": "object",
                                                "description": "Organization data",
                                                "properties": map[string]interface{}{
                                                        "id": map[string]string{"type": "string"},
                                                        "name": map[string]string{"type": "string"},
                                                        "displayName": map[string]string{"type": "string"},
                                                        "type": map[string]string{"type": "string"},
                                                        "links": map[string]string{"type": "object"},
                                                },
                                                "required": []string{"id", "name", "displayName", "type"},
                                        },
                                        {
                                                "type": "object",
                                                "description": "Team data with members",
                                                "properties": map[string]interface{}{
                                                        "teamId": map[string]string{"type": "string"},
                                                        "displayName": map[string]string{"type": "string"},
                                                        "description": map[string]string{"type": "string"},
                                                        "teamType": map[string]string{"type": "string"},
                                                        "organizationId": map[string]string{"type": "string"},
                                                        "creatorId": map[string]string{"type": "string"},
                                                        "members": map[string]interface{}{
                                                                "type": "array",
                                                                "items": map[string]interface{}{
                                                                        "type": "object",
                                                                        "properties": map[string]interface{}{
                                                                                "accountId": map[string]string{"type": "string"},
                                                                                "email": map[string]string{"type": "string"},
                                                                                "displayName": map[string]string{"type": "string"},
                                                                                "accountType": map[string]string{"type": "string"},
                                                                        },
                                                                        "required": []string{"accountId", "displayName", "accountType"},
                                                                },
                                                        },
                                                        "links": map[string]string{"type": "object"},
                                                },
                                                "required": []string{"teamId", "displayName", "teamType", "organizationId"},
                                        },
                                        {
                                                "type": "object",
                                                "description": "Processing statistics",
                                                "properties": map[string]interface{}{
                                                        "startTime": map[string]string{"type": "string", "format": "date-time"},
                                                        "organizationsFound": map[string]string{"type": "integer"},
                                                        "teamsFound": map[string]string{"type": "integer"},
                                                        "teamsProcessed": map[string]string{"type": "integer"},
                                                        "membersFound": map[string]string{"type": "integer"},
                                                        "apiCallsTotal": map[string]string{"type": "integer"},
                                                        "apiCallsSucceeded": map[string]string{"type": "integer"},
                                                        "apiCallsFailed": map[string]string{"type": "integer"},
                                                        "retryAttempts": map[string]string{"type": "integer"},
                                                        "processingTime": map[string]string{"type": "string"},
                                                        "status": map[string]interface{}{
                                                                "type": "string",
                                                                "enum": []string{"processing", "completed", "failed"},
                                                        },
                                                        "lastUpdated": map[string]string{"type": "string", "format": "date-time"},
                                                },
                                                "required": []string{"startTime", "status", "lastUpdated"},
                                        },
                                },
                        },
                },
        }
}

// NewAtlassianClient creates a new Atlassian API client
func NewAtlassianClient(config Config) *AtlassianClient {
        return &AtlassianClient{
                config: config,
                httpClient: &http.Client{
                        Timeout: 30 * time.Second,
                },
        }
}

// makeRequest performs an HTTP request with retry logic and proper authentication
func (c *AtlassianClient) makeRequest(method, url string, bodyData []byte, useOAuth bool, stats *StatisticsTracker) ([]byte, error) {
        stats.stats.APICallsTotal++
        var lastErr error
        maxRetries := 5 // Increased retry budget
        for attempt := 0; attempt < maxRetries; attempt++ {
                if attempt > 0 {
                        stats.stats.RetryAttempts++
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
                                        // Cap retry delay at 120 seconds
                                        if delay > 120*time.Second {
                                                delay = 120 * time.Second
                                        }
                                        time.Sleep(delay)
                                } else {
                                        // Try parsing as HTTP-date (multiple formats)
                                        var httpDate time.Time
                                        var err error
                                        // RFC1123: "Mon, 02 Jan 2006 15:04:05 GMT"
                                        if httpDate, err = time.Parse(time.RFC1123, retryAfter); err != nil {
                                                // RFC850: "Monday, 02-Jan-06 15:04:05 GMT"
                                                if httpDate, err = time.Parse(time.RFC850, retryAfter); err != nil {
                                                        // ANSIC: "Mon Jan _2 15:04:05 2006"
                                                        httpDate, err = time.Parse(time.ANSIC, retryAfter)
                                                }
                                        }
                                        if err == nil {
                                                delay := time.Until(httpDate)
                                                if delay > 0 {
                                                        // Cap at 120 seconds but honor the server guidance
                                                        if delay > 120*time.Second {
                                                                delay = 120 * time.Second
                                                        }
                                                        time.Sleep(delay)
                                                }
                                        }
                                }
                        } else {
                                // No Retry-After header, use exponential backoff with jitter
                                baseDelay := time.Duration(attempt*attempt) * time.Second
                                jitter := time.Duration(attempt*100) * time.Millisecond
                                time.Sleep(baseDelay + jitter)
                        }
                        lastErr = fmt.Errorf("rate limited, attempt %d/%d", attempt+1, maxRetries)
                        continue
                }

                // Retry on 5xx server errors
                if resp.StatusCode >= 500 {
                        fmt.Fprintf(os.Stderr, "Server error %d, attempt %d/%d\n", resp.StatusCode, attempt+1, maxRetries)
                        // Exponential backoff for server errors
                        delay := time.Duration(attempt*attempt) * time.Second
                        if delay > 30*time.Second {
                                delay = 30 * time.Second
                        }
                        time.Sleep(delay)
                        lastErr = fmt.Errorf("server error %d, attempt %d/%d: %s", resp.StatusCode, attempt+1, maxRetries, string(respBody))
                        continue
                }

                // Non-retryable client errors
                if resp.StatusCode != http.StatusOK {
                        stats.stats.APICallsFailed++
                        return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
                }

                stats.stats.APICallsSucceeded++
                return respBody, nil
        }

        stats.stats.APICallsFailed++
        return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// GetOrganization retrieves organization information using admin API
func (c *AtlassianClient) GetOrganization(stats *StatisticsTracker) (*Organization, error) {
        url := fmt.Sprintf("%s/admin/v1/orgs/%s", c.config.AdminBaseURL, c.config.OrgID)

        body, err := c.makeRequest("GET", url, nil, true, stats) // Use OAuth
        if err != nil {
                return nil, fmt.Errorf("getting organization: %w", err)
        }

        var org Organization
        if err := json.Unmarshal(body, &org); err != nil {
                return nil, fmt.Errorf("unmarshaling organization: %w", err)
        }

        stats.stats.OrganizationsFound++
        return &org, nil
}

// GetTeams retrieves all teams in the organization with streaming pagination
func (c *AtlassianClient) GetTeams(writer io.Writer, stats *StatisticsTracker) error {
        nextCursor := ""
        teamsProcessed := 0

        for {
                // Use GET method for listing teams with query parameters
                teamsURL := fmt.Sprintf("https://api.atlassian.com/public/teams/v1/org/%s/teams", c.config.OrgID)
                
                // Add cursor parameter if we have one
                if nextCursor != "" {
                        teamsURL += "?cursor=" + url.QueryEscape(nextCursor)
                }

                respBody, err := c.makeRequest("GET", teamsURL, nil, false, stats) // Use basic auth, no body
                if err != nil {
                        return fmt.Errorf("getting teams: %w", err)
                }

                // Teams list response structure (based on API docs)
                var response struct {
                        Cursor   string `json:"cursor"`
                        Entities []Team `json:"entities"`
                }

                if err := json.Unmarshal(respBody, &response); err != nil {
                        return fmt.Errorf("unmarshaling teams: %w", err)
                }

                // Update teams found count
                stats.stats.TeamsFound += len(response.Entities)
                
                // Stream teams as we get them - don't accumulate in memory
                for _, team := range response.Entities {
                        // Set Name field for internal use (since we removed the duplicate JSON tag)
                        team.Name = team.DisplayName
                        fmt.Fprintf(os.Stderr, "Processing team: %s\n", team.DisplayName)
                        
                        // Get team members
                        members, err := c.GetTeamMembers(team.ID, stats)
                        if err != nil {
                                fmt.Fprintf(os.Stderr, "Warning: Error fetching members for team %s: %v\n", team.Name, err)
                                members = []Member{} // Continue with empty members list
                        }
                        
                        team.Members = members
                        stats.stats.MembersFound += len(members)
                        
                        if err := outputNDJSON(writer, "team", team); err != nil {
                                return fmt.Errorf("outputting team: %w", err)
                        }

                        stats.stats.TeamsProcessed++
                        teamsProcessed++
                        
                        // Emit periodic statistics every 10 teams
                        if teamsProcessed%10 == 0 {
                                if err := outputNDJSON(writer, "statistics", stats.UpdateStats()); err != nil {
                                        fmt.Fprintf(os.Stderr, "Warning: Failed to output statistics: %v\n", err)
                                }
                        }
                        
                        // Rate limiting - be nice to the API
                        time.Sleep(200 * time.Millisecond)
                }

                // Check if we have more pages (cursor-based pagination)
                if response.Cursor == "" {
                        break
                }
                nextCursor = response.Cursor
        }

        fmt.Fprintf(os.Stderr, "Successfully processed %d teams\n", teamsProcessed)
        return nil
}

// GetTeamMembers retrieves members for a specific team with pagination
func (c *AtlassianClient) GetTeamMembers(teamID string, stats *StatisticsTracker) ([]Member, error) {
        var allMembers []Member
        nextCursor := ""

        for {
                url := fmt.Sprintf("https://api.atlassian.com/public/teams/v1/org/%s/teams/%s/members", c.config.OrgID, teamID)

                // Prepare request body for pagination
                reqBody := map[string]interface{}{
                        "first": 50, // Max items per page
                }
                if nextCursor != "" {
                        reqBody["after"] = nextCursor
                }

                bodyBytes, err := json.Marshal(reqBody)
                if err != nil {
                        return nil, fmt.Errorf("marshaling request body: %w", err)
                }

                respBody, err := c.makeRequest("POST", url, bodyBytes, false, stats) // Use basic auth
                if err != nil {
                        return nil, fmt.Errorf("getting team members: %w", err)
                }

                // Team members response uses different structure
                var response struct {
                        Results []Member `json:"results"`
                        PageInfo struct {
                                HasNextPage bool   `json:"hasNextPage"`
                                EndCursor   string `json:"endCursor"`
                        } `json:"pageInfo"`
                }

                if err := json.Unmarshal(respBody, &response); err != nil {
                        return nil, fmt.Errorf("unmarshaling team members: %w", err)
                }

                allMembers = append(allMembers, response.Results...)

                if !response.PageInfo.HasNextPage {
                        break
                }
                nextCursor = response.PageInfo.EndCursor
        }

        return allMembers, nil
}

// outputNDJSON writes an object as NDJSON to the specified writer
func outputNDJSON(writer io.Writer, objType string, data interface{}) error {
        output := NDJSONOutput{
                Type:      objType,
                Timestamp: time.Now().UTC().Format(time.RFC3339),
                Data:      data,
        }

        jsonData, err := json.Marshal(output)
        if err != nil {
                return fmt.Errorf("marshaling NDJSON: %w", err)
        }

        _, err = fmt.Fprintln(writer, string(jsonData))
        return err
}


func main() {
        var (
                adminBaseURL = flag.String("admin-base-url", "https://api.atlassian.com", "Atlassian Admin API base URL")
                accessToken  = flag.String("access-token", "", "OAuth 2.0 access token for admin APIs")
                username     = flag.String("username", "", "Atlassian username/email for teams API")
                apiToken     = flag.String("api-token", "", "Atlassian API token for teams API")
                orgID        = flag.String("org-id", "", "Atlassian organization ID")
                output       = flag.String("output", "", "Output file (default: stdout)")
        )
        flag.Parse()

        // Validate required parameters
        if *orgID == "" {
                fmt.Fprintf(os.Stderr, "Error: org-id is required\n")
                flag.Usage()
                os.Exit(1)
        }

        // Check that we have either access token (for admin API) or username+token (for teams API)
        if *accessToken == "" && (*username == "" || *apiToken == "") {
                fmt.Fprintf(os.Stderr, "Error: Either access-token (for admin API) or username+api-token (for teams API) are required\n")
                flag.Usage()
                os.Exit(1)
        }

        // Set up output writer
        var writer io.Writer = os.Stdout
        if *output != "" {
                file, err := os.Create(*output)
                if err != nil {
                        log.Fatalf("Error creating output file: %v", err)
                }
                defer file.Close()
                writer = file
        }

        // Initialize client
        config := Config{
                AdminBaseURL: strings.TrimSuffix(*adminBaseURL, "/"),
                AccessToken:  *accessToken,
                Username:     *username,
                APIToken:     *apiToken,
                OrgID:        *orgID,
        }

        client := NewAtlassianClient(config)

        // Initialize statistics tracker
        stats := NewStatisticsTracker()

        // Output JSON schema as the first element
        schema := createJSONSchema()
        if err := outputNDJSON(writer, "schema", schema); err != nil {
                log.Fatalf("Error outputting JSON schema: %v", err)
        }

        // Define error variable and finalization function to ensure it always runs
        var processingErr error
        defer func() {
                // Set final status based on errors
                if processingErr != nil {
                        stats.stats.Status = "failed"
                } else {
                        stats.stats.Status = "completed"
                }

                // Always output final statistics
                finalStats := stats.UpdateStats()
                if err := outputNDJSON(writer, "statistics", finalStats); err != nil {
                        fmt.Fprintf(os.Stderr, "Warning: Failed to output final statistics: %v\n", err)
                }

                // Always output human-readable statistics to stderr
                fmt.Fprintf(os.Stderr, "\n=== Processing Summary ===\n")
                fmt.Fprintf(os.Stderr, "Organizations found: %d\n", finalStats.OrganizationsFound)
                fmt.Fprintf(os.Stderr, "Teams found: %d\n", finalStats.TeamsFound)
                fmt.Fprintf(os.Stderr, "Teams processed: %d\n", finalStats.TeamsProcessed)
                fmt.Fprintf(os.Stderr, "Members found: %d\n", finalStats.MembersFound)
                fmt.Fprintf(os.Stderr, "API calls made: %d (succeeded: %d, failed: %d)\n", 
                        finalStats.APICallsTotal, finalStats.APICallsSucceeded, finalStats.APICallsFailed)
                fmt.Fprintf(os.Stderr, "Retry attempts: %d\n", finalStats.RetryAttempts)
                fmt.Fprintf(os.Stderr, "Total processing time: %s\n", finalStats.ProcessingTime)
                fmt.Fprintf(os.Stderr, "Status: %s\n", finalStats.Status)
                if *output != "" {
                        fmt.Fprintf(os.Stderr, "Output written to: %s\n", *output)
                }

                // Exit with proper code if there was an error
                if processingErr != nil {
                        os.Exit(1)
                }
        }()

        // Get organization information (only if we have access token)
        if config.AccessToken != "" {
                fmt.Fprintf(os.Stderr, "Fetching organization information...\n")
                org, err := client.GetOrganization(stats)
                if err != nil {
                        fmt.Fprintf(os.Stderr, "Warning: Could not fetch organization info (requires OAuth access token): %v\n", err)
                } else {
                        if err := outputNDJSON(writer, "organization", org); err != nil {
                                processingErr = fmt.Errorf("error outputting organization: %w", err)
                                fmt.Fprintf(os.Stderr, "Error: %v\n", processingErr)
                                return
                        }
                }
        }

        // Get teams (requires username + API token)
        if config.Username != "" && config.APIToken != "" {
                fmt.Fprintf(os.Stderr, "Fetching teams...\n")
                if err := client.GetTeams(writer, stats); err != nil {
                        processingErr = fmt.Errorf("error fetching teams: %w", err)
                        fmt.Fprintf(os.Stderr, "Error: %v\n", processingErr)
                        return
                }
        } else {
                fmt.Fprintf(os.Stderr, "Skipping teams (requires username and api-token)\n")
        }
}