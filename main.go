package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"atlassian-org-tool/internal/client"
	"atlassian-org-tool/internal/schema"
	"atlassian-org-tool/internal/stats"
	"atlassian-org-tool/internal/types"
)

// outputNDJSON outputs data in NDJSON format
func outputNDJSON(writer io.Writer, dataType string, data interface{}) error {
	output := types.NDJSONOutput{
		Type:      dataType,
		Timestamp: time.Now().Format(time.RFC3339),
		Data:      data,
	}

	encoder := json.NewEncoder(writer)
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("encoding NDJSON: %w", err)
	}

	return nil
}

func main() {
	// Command line flags
	adminBaseURL := flag.String("admin-base-url", "https://api.atlassian.com", "Atlassian Admin API base URL")
	accessToken := flag.String("access-token", "", "OAuth 2.0 access token for admin APIs")
	username := flag.String("username", "", "Atlassian username/email for teams API")
	apiToken := flag.String("api-token", "", "Atlassian API token for teams API")
	orgID := flag.String("org-id", "", "Atlassian organization ID")
	output := flag.String("output", "", "Output file (default: stdout)")

	flag.Parse()

	// Validate required parameters
	if *orgID == "" {
		log.Fatal("org-id is required")
	}

	// Must have either access token OR username+api-token
	hasOAuth := *accessToken != ""
	hasBasicAuth := *username != "" && *apiToken != ""

	if !hasOAuth && !hasBasicAuth {
		log.Fatal("Must provide either -access-token (for admin API) or both -username and -api-token (for teams API)")
	}

	// Setup output writer
	var writer io.Writer = os.Stdout
	if *output != "" {
		file, err := os.Create(*output)
		if err != nil {
			log.Fatalf("Error creating output file: %v", err)
		}
		defer file.Close()
		writer = file
		fmt.Fprintf(os.Stderr, "Writing output to: %s\n", *output)
	}

	// Create configuration
	config := types.Config{
		AdminBaseURL: *adminBaseURL,
		AccessToken:  *accessToken,
		Username:     *username,
		APIToken:     *apiToken,
		OrgID:        *orgID,
	}

	// Create API client and statistics tracker
	apiClient := client.NewAtlassianClient(config)
	tracker := stats.NewTracker()

	// Output JSON schema as the first element
	jsonSchema := schema.CreateJSONSchema()
	if err := outputNDJSON(writer, "schema", jsonSchema); err != nil {
		log.Fatalf("Error outputting JSON schema: %v", err)
	}

	// Define error variable and finalization function to ensure it always runs
	var processingErr error
	defer func() {
		// Set final status based on errors
		if processingErr != nil {
			tracker.SetStatus("failed")
		} else {
			tracker.SetStatus("completed")
		}

		// Always output final statistics
		finalStats := tracker.UpdateStats()
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
		org, err := apiClient.GetOrganization(tracker)
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
		teams, err := apiClient.GetTeams(tracker)
		if err != nil {
			processingErr = fmt.Errorf("error fetching teams: %w", err)
			fmt.Fprintf(os.Stderr, "Error: %v\n", processingErr)
			return
		}

		fmt.Fprintf(os.Stderr, "Found %d teams, fetching members...\n", len(teams))

		// Process each team and get its members
		for i, team := range teams {
			fmt.Fprintf(os.Stderr, "Processing team %d/%d: %s\n", i+1, len(teams), team.DisplayName)

			// Get team members
			members, err := apiClient.GetTeamMembers(team.ID, tracker)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Could not fetch members for team %s: %v\n", team.DisplayName, err)
				// Continue with empty members list
				members = []types.Member{}
			}

			// Add members to team
			team.Members = members
			tracker.IncrementTeamsProcessed()

			// Output team with members
			if err := outputNDJSON(writer, "team", team); err != nil {
				processingErr = fmt.Errorf("error outputting team %s: %w", team.DisplayName, err)
				fmt.Fprintf(os.Stderr, "Error: %v\n", processingErr)
				return
			}

			// Output periodic statistics every 10 teams
			if (i+1)%10 == 0 {
				periodicStats := tracker.UpdateStats()
				if err := outputNDJSON(writer, "statistics", periodicStats); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to output periodic statistics: %v\n", err)
				}
			}
		}
	}

	fmt.Fprintf(os.Stderr, "Processing completed successfully.\n")
}
