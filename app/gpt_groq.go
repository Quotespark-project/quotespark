package main

import (
	"context"
	"encoding/json"
	"os"
	"fmt"
	groq "github.com/conneroisu/groq-go"
)

var groqClient *groq.Client

func initGroq() {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		panic("GROQ_API_KEY not set")
	}
	client, err := groq.NewClient(apiKey)
	if err != nil {
		panic(err)
	}
	groqClient = client
}

func getDailyQuestionGroq() (string, error) {
	req := groq.ChatCompletionRequest{
		Model: "llama3-70b-8192", // or another Groq-supported model
		Messages: []groq.ChatCompletionMessage{
			{Role: "user", Content: "One thought-provoking motivational question for today. Be concise and short. The question should be in second person"},
		},
	}
	resp, err := groqClient.ChatCompletion(context.Background(), req)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response")
	}
	return resp.Choices[0].Message.Content, nil
}

func generateDummyReflections(question string) ([]string, error) {
	req := groq.ChatCompletionRequest{
		Model: "llama3-70b-8192",
		Messages: []groq.ChatCompletionMessage{
			{
				Role:    "user",
				Content: fmt.Sprintf(`The journaling question is: "%s". Generate 10 diverse, anonymous, emotional reflections answering this question. Each reflection should be about 1-2 sentences. Output them as a JSON list of strings.`, question),
			},
		},
	}
	resp, err := groqClient.ChatCompletion(context.Background(), req)
	if err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no GPT response")
	}

	var reflections []string
	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &reflections)
	if err != nil {
		return nil, fmt.Errorf("failed to parse GPT response: %w", err)
	}

	return reflections, nil
}
