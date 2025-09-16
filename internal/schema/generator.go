package schema

import "atlassian-org-tool/internal/types"

// CreateJSONSchema creates a JSON schema for the NDJSON output
func CreateJSONSchema() types.JSONSchema {
	return types.JSONSchema{
		Schema:      "https://json-schema.org/draft/2020-12/schema",
		Title:       "Atlassian Organization and Team Data Stream",
		Description: "Schema for NDJSON stream containing Atlassian organization and team metadata",
		Type:        "object",
		Properties: map[string]interface{}{
			"type": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"schema", "organization", "team", "statistics"},
				"description": "Type of data in this line",
			},
			"timestamp": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "ISO 8601 timestamp when this data was generated",
			},
			"data": map[string]interface{}{
				"oneOf": []map[string]interface{}{
					{
						"type":        "object",
						"description": "Organization data",
						"properties": map[string]interface{}{
							"id":          map[string]string{"type": "string"},
							"name":        map[string]string{"type": "string"},
							"displayName": map[string]string{"type": "string"},
							"type":        map[string]string{"type": "string"},
							"links":       map[string]string{"type": "object"},
						},
						"required": []string{"id", "name", "displayName", "type"},
					},
					{
						"type":        "object",
						"description": "Team data with members",
						"properties": map[string]interface{}{
							"teamId":         map[string]string{"type": "string"},
							"displayName":    map[string]string{"type": "string"},
							"description":    map[string]string{"type": "string"},
							"teamType":       map[string]string{"type": "string"},
							"organizationId": map[string]string{"type": "string"},
							"creatorId":      map[string]string{"type": "string"},
							"links":          map[string]string{"type": "object"},
							"members": map[string]interface{}{
								"type": "array",
								"items": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"accountId":   map[string]string{"type": "string"},
										"email":       map[string]string{"type": "string"},
										"displayName": map[string]string{"type": "string"},
										"accountType": map[string]string{"type": "string"},
									},
									"required": []string{"accountId", "displayName", "accountType"},
								},
							},
						},
						"required": []string{"teamId", "displayName", "teamType", "organizationId"},
					},
					{
						"type":        "object",
						"description": "Processing statistics",
						"properties": map[string]interface{}{
							"startTime":          map[string]interface{}{"type": "string", "format": "date-time"},
							"organizationsFound": map[string]string{"type": "integer"},
							"teamsFound":         map[string]string{"type": "integer"},
							"teamsProcessed":     map[string]string{"type": "integer"},
							"membersFound":       map[string]string{"type": "integer"},
							"apiCallsTotal":      map[string]string{"type": "integer"},
							"apiCallsSucceeded":  map[string]string{"type": "integer"},
							"apiCallsFailed":     map[string]string{"type": "integer"},
							"retryAttempts":      map[string]string{"type": "integer"},
							"processingTime":     map[string]string{"type": "string"},
							"status": map[string]interface{}{
								"type": "string",
								"enum": []string{"processing", "completed", "failed"},
							},
							"lastUpdated": map[string]interface{}{"type": "string", "format": "date-time"},
						},
						"required": []string{"startTime", "status", "lastUpdated"},
					},
				},
			},
		},
		Required: []string{"type", "timestamp", "data"},
	}
}
