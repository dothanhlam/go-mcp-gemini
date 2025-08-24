package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"google.golang.org/api/aiplatform/v1beta1"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		log.Fatal("GOOGLE_CLOUD_PROJECT environment variable must be set")
	}

	location := "us-central1"
	modelID := "gemini-1.5-flash-001"

	endpoint := fmt.Sprintf("https://%s-aiplatform.googleapis.com", location)
	client, err := aiplatform.NewService(ctx, option.WithEndpoint(endpoint))
	if err != nil {
		log.Fatalf("Failed to create AI Platform client: %v", err)
	}

	model := fmt.Sprintf("projects/%s/locations/%s/publishers/google/models/%s", projectID, location, modelID)

	tools := []*aiplatform.Tool{
		{
			FunctionDeclarations: []*aiplatform.FunctionDeclaration{
				{
					Name:        "get_current_weather",
					Description: "Get the current weather in a given location",
					Parameters: &aiplatform.Schema{
						Type: "OBJECT",
						Properties: map[string]*aiplatform.Schema{
							"location": {Type: "STRING", Description: "The city and state, e.g. San Francisco, CA"},
						},
						Required: []string{"location"},
					},
				},
			},
		},
	}

	var conversationHistory []*aiplatform.Content
	prompt := "What is the weather like in Boston?"
	conversationHistory = append(conversationHistory, &aiplatform.Content{
		Role: "user",
		Parts: []*aiplatform.Part{
			{Text: prompt},
		},
	})

	for {
		req := &aiplatform.GenerateContentRequest{
			Contents: conversationHistory,
			Tools:    tools,
		}

		resp, err := client.Projects.Locations.Publishers.Models.GenerateContent(model, req).Do()
		if err != nil {
			log.Fatalf("Failed to generate content: %v", err)
		}

		if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
			log.Println("No response from model")
			break
		}

		candidate := resp.Candidates[0]
		part := candidate.Content.Parts[0]

		if part.FunctionCall != nil {
			conversationHistory = append(conversationHistory, candidate.Content)
			functionCall := part.FunctionCall
			toolResponse := dispatch(functionCall)
			conversationHistory = append(conversationHistory, toolResponse)
		} else if part.Text != "" {
			fmt.Println(part.Text)
			break
		}
	}
}

func dispatch(call *aiplatform.FunctionCall) *aiplatform.Content {
	var toolResponse *aiplatform.Content
	switch call.Name {
	case "get_current_weather":
		var args struct {
			Location string `json:"location"`
		}
		if err := json.Unmarshal(call.Args, &args); err != nil {
			log.Fatalf("Failed to unmarshal arguments: %v", err)
		}
		result := getCurrentWeather(args.Location)
		resultBytes, _ := json.Marshal(result)
		toolResponse = &aiplatform.Content{
			Role: "function",
			Parts: []*aiplatform.Part{
				{
					FunctionResponse: &aiplatform.FunctionResponse{
						Name:     "get_current_weather",
						Response: resultBytes,
					},
				},
			},
		}
	return toolResponse
}

func getCurrentWeather(location string) map[string]interface{} {
	return map[string]interface{}{
		"location":    location,
		"temperature": "72",
		"unit":        "fahrenheit",
		"forecast":    "sunny",
	}
}
