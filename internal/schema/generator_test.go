package schema

import (
	"encoding/json"
	"testing"
)

func TestCreateJSONSchema(t *testing.T) {
	schema := CreateJSONSchema()

	// Test basic schema structure
	if schema.Schema != "https://json-schema.org/draft/2020-12/schema" {
		t.Errorf("Expected schema URL 'https://json-schema.org/draft/2020-12/schema', got '%s'", schema.Schema)
	}

	if schema.Title == "" {
		t.Error("Expected non-empty title")
	}

	if schema.Description == "" {
		t.Error("Expected non-empty description")
	}

	if schema.Type != "object" {
		t.Errorf("Expected type 'object', got '%s'", schema.Type)
	}

	// Test required fields
	expectedRequired := []string{"type", "timestamp", "data"}
	if len(schema.Required) != len(expectedRequired) {
		t.Errorf("Expected %d required fields, got %d", len(expectedRequired), len(schema.Required))
	}

	for i, expected := range expectedRequired {
		if i >= len(schema.Required) || schema.Required[i] != expected {
			t.Errorf("Expected required field '%s' at index %d, got '%s'", expected, i, schema.Required[i])
		}
	}

	// Test properties exist
	if schema.Properties == nil {
		t.Fatal("Properties should not be nil")
	}

	// Check that key properties exist
	if _, exists := schema.Properties["type"]; !exists {
		t.Error("Expected 'type' property to exist")
	}

	if _, exists := schema.Properties["timestamp"]; !exists {
		t.Error("Expected 'timestamp' property to exist")
	}

	if _, exists := schema.Properties["data"]; !exists {
		t.Error("Expected 'data' property to exist")
	}
}

func TestJSONSchemaMarshaling(t *testing.T) {
	schema := CreateJSONSchema()

	// Test that the schema can be marshaled to JSON
	data, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("Failed to marshal schema to JSON: %v", err)
	}

	// Test that it can be unmarshaled back
	var unmarshaled map[string]interface{}
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal schema JSON: %v", err)
	}

	// Verify key fields are present in unmarshaled data
	if unmarshaled["$schema"] != schema.Schema {
		t.Errorf("Schema field mismatch after marshal/unmarshal")
	}

	if unmarshaled["title"] != schema.Title {
		t.Errorf("Title field mismatch after marshal/unmarshal")
	}

	if unmarshaled["type"] != schema.Type {
		t.Errorf("Type field mismatch after marshal/unmarshal")
	}
}

func TestSchemaDataOneOfStructure(t *testing.T) {
	schema := CreateJSONSchema()

	// Navigate to the data.oneOf structure
	dataProperty, exists := schema.Properties["data"]
	if !exists {
		t.Fatal("data property should exist")
	}

	dataMap, ok := dataProperty.(map[string]interface{})
	if !ok {
		t.Fatal("data property should be a map")
	}

	oneOf, exists := dataMap["oneOf"]
	if !exists {
		t.Fatal("data.oneOf should exist")
	}

	oneOfSlice, ok := oneOf.([]map[string]interface{})
	if !ok {
		t.Fatal("data.oneOf should be a slice of maps")
	}

	// Should have at least 3 schemas: organization, team, statistics
	if len(oneOfSlice) < 3 {
		t.Errorf("Expected at least 3 schemas in oneOf, got %d", len(oneOfSlice))
	}

	// Check that each schema in oneOf has the expected structure
	for i, schema := range oneOfSlice {
		if schema["type"] != "object" {
			t.Errorf("Schema %d should have type 'object'", i)
		}

		if _, exists := schema["description"]; !exists {
			t.Errorf("Schema %d should have a description", i)
		}

		if _, exists := schema["properties"]; !exists {
			t.Errorf("Schema %d should have properties", i)
		}
	}
}
