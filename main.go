package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/vertexai/genai"
)

func main() {
	ctx := context.Background()
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		log.Fatal("GOOGLE_CLOUD_PROJECT environment variable must be set")
	}

	location := "us-central1"
	modelID := "gemini-1.5-flash-001"

	client, err := genai.NewClient(ctx, projectID, location)
	if err != nil {
		log.Fatalf("Failed to create new client: %v", err)
	}
	defer client.Close()

	model := client.GenerativeModel(modelID)
	model.Tools = []*genai.Tool{
		{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				{
					Name:        "get_current_weather",
					Description: "Get the current weather in a given location",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"location": {
								Type:        genai.TypeString,
								Description: "The city and state, e.g. San Francisco, CA",
							},
						},
						Required: []string{"location"},
					},
				},
			},
		},
	}

	prompt := "What is the weather like in Boston?"
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Fatalf("Failed to generate content: %v", err)
	}

	part := resp.Candidates[0].Content.Parts[0]
	if fc, ok := part.(genai.FunctionCall); ok {
		fmt.Printf("Function call: %s\n", fc.Name)
		for name, val := range fc.Args {
			fmt.Printf("  Arg: %s = %v\n", name, val)
		}
	} else {
		fmt.Println("No function call requested.")
	}
}