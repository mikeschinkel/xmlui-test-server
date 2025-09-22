package cfgldr

import (
	jsonv2 "encoding/json/v2"
	"fmt"
)

func example() {
	// Example JSON with array params
	jsonArrayParams := `{
		"endpoint": "GET /users/{id:int}",
		"description": "Get user by ID",
		"query": "SELECT * FROM users WHERE id = :id",
		"cardinality": "one",
		"row_type": "json",
		"params": [
			{"name": "id", "type": "int", "constraints": ""}
		]
	}`

	// Example JSON with map params
	jsonMapParams := `{
		"endpoint": "GET /posts/{slug:slug:length[5..50]}",
		"description": "Get post by slug",
		"query": "SELECT * FROM posts WHERE slug = :slug",
		"cardinality": "one", 
		"row_type": "json",
		"params": {
			"slug": "slug:length[5..50]"
		}
	}`

	fmt.Println("=== JSON v2 with Inline Implementation ===")

	// Test with array params
	var endpoint1 APIEndpointV2
	err := jsonv2.Unmarshal([]byte(jsonArrayParams), &endpoint1)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Array params: %+v\n", endpoint1)

	// Test with map params
	var endpoint2 APIEndpointV2
	err = jsonv2.Unmarshal([]byte(jsonMapParams), &endpoint2)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Map params: %+v\n", endpoint2)

	// Test marshaling back to JSON
	output, err := jsonv2.Marshal(endpoint2)
	if err != nil {
		fmt.Printf("Marshal error: %v\n", err)
		return
	}
	fmt.Printf("Marshaled back: %s\n", output)
}
