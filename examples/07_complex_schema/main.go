// Package main demonstrates complex nested structured outputs with OpenAI agents.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	agents "github.com/MitulShah1/openai-agents-go"
	"github.com/MitulShah1/openai-agents-go/internal/jsonschema"
)

// Person represents a person with contact information
type Person struct {
	Name    string   `json:"name"`
	Age     int      `json:"age"`
	Email   string   `json:"email"`
	Address Address  `json:"address"`
	Hobbies []string `json:"hobbies"`
}

// Address represents a physical address
type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	ZipCode string `json:"zipCode"`
}

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable not set")
	}

	// Initialize OpenAI client
	client := openai.NewClient(option.WithAPIKey(apiKey))
	runner := agents.NewRunner(&client)

	// Define complex nested JSON schema
	addressSchema := jsonschema.Object().
		WithDescription("Physical address").
		WithProperty("street", jsonschema.String().WithDescription("Street address")).
		WithProperty("city", jsonschema.String().WithDescription("City name")).
		WithProperty("state", jsonschema.String().
			WithDescription("State or province").
			WithMinLength(2).WithMaxLength(2)).
		WithProperty("zipCode", jsonschema.String().
			WithDescription("ZIP/Postal code").
			WithPattern("^\\d{5}$")).
		WithRequired("street", "city", "state", "zipCode")

	personSchema := jsonschema.Object().
		WithDescription("Person with contact information").
		WithProperty("name", jsonschema.String().
			WithDescription("Full name")).
		WithProperty("age", jsonschema.Integer().
			WithDescription("Age in years").
			WithMinimum(0).WithMaximum(150)).
		WithProperty("email", jsonschema.String().
			WithDescription("Email address").
			WithPattern("^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$")).
		WithProperty("address", addressSchema).
		WithProperty("hobbies", jsonschema.Array(jsonschema.String()).
			WithDescription("List of hobbies")).
		WithRequired("name", "age", "email", "address", "hobbies")

	// Create agent with complex structured output
	agent := agents.NewAgent("Data Extractor")
	agent.Instructions = "You are a data extraction expert. Extract structured information from text."
	agent.ResponseFormat = jsonschema.JSONSchema("person_info", personSchema).
		WithDescription("Extract person information from text")

	// Text to extract from
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage(`Extract the following information:
John Smith is a 32-year-old software engineer living at 
123 Main Street, Boston, MA 02101. You can reach him at 
john.smith@example.com. He enjoys hiking, photography, and cooking.`),
	}

	ctx := context.Background()
	result, err := runner.Run(ctx, agent, messages, nil, nil, nil, "")
	if err != nil {
		log.Fatalf("Error running agent: %v", err)
	}

	// Parse the structured response
	if len(result.Messages) > 0 {
		// Get the final output which should be structured JSON
		if result.FinalOutput != "" {
			var person Person
			if err := json.Unmarshal([]byte(result.FinalOutput), &person); err != nil {
				log.Fatalf("Error parsing response: %v", err)
			}

			fmt.Println("Extracted Person Information:")
			fmt.Println("============================")
			fmt.Printf("\nName: %s\n", person.Name)
			fmt.Printf("Age: %d\n", person.Age)
			fmt.Printf("Email: %s\n", person.Email)
			fmt.Printf("\nAddress:\n")
			fmt.Printf("  Street: %s\n", person.Address.Street)
			fmt.Printf("  City: %s\n", person.Address.City)
			fmt.Printf("  State: %s\n", person.Address.State)
			fmt.Printf("  ZIP: %s\n", person.Address.ZipCode)
			fmt.Printf("\nHobbies:\n")
			for _, hobby := range person.Hobbies {
				fmt.Printf("  - %s\n", hobby)
			}
		}
	}

	fmt.Printf("\nTokens used: %d\n", result.Usage.TotalTokens)
}
