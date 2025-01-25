package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type ChatResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
}

func main() {
	// Load environment variables
	endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
	apiKey := os.Getenv("AZURE_OPENAI_API_KEY")
	model := os.Getenv("AZURE_OPENAI_MODEL") // E.g., "gpt-4"

	if endpoint == "" || apiKey == "" || model == "" {
		fmt.Println("Please set AZURE_OPENAI_ENDPOINT, AZURE_OPENAI_API_KEY, and AZURE_OPENAI_MODEL environment variables.")
		return
	}

	// Base URL for Azure OpenAI
	url := fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=2023-06-01-preview", endpoint, model)

	// Initialize chat history
	var chatHistory []ChatMessage
	chatHistory = append(chatHistory, ChatMessage{Role: "system", Content: "You are a helpful assistant."})

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Start chatting with the Azure OpenAI model (type 'exit' to quit):")

	for {
		fmt.Print("You: ")
		userInput, _ := reader.ReadString('\n')
		userInput = userInput[:len(userInput)-1] // Remove newline character

		if userInput == "exit" {
			fmt.Println("Exiting chat. Goodbye!")
			break
		}

		// Add user input to chat history
		chatHistory = append(chatHistory, ChatMessage{Role: "user", Content: userInput})

		// Prepare request payload
		requestBody := ChatRequest{
			Model:    model,
			Messages: chatHistory,
		}
		jsonBody, err := json.Marshal(requestBody)
		if err != nil {
			fmt.Println("Error creating request payload:", err)
			continue
		}

		// Make HTTP request
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
		if err != nil {
			fmt.Println("Error creating request:", err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("api-key", apiKey)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error making API request:", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Printf("API request failed with status %d: %s\n", resp.StatusCode, string(body))
			continue
		}

		// Parse response
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			continue
		}

		var chatResponse ChatResponse
		if err := json.Unmarshal(body, &chatResponse); err != nil {
			fmt.Println("Error parsing response JSON:", err)
			continue
		}

		// Display assistant response
		if len(chatResponse.Choices) > 0 {
			assistantMessage := chatResponse.Choices[0].Message
			fmt.Printf("Assistant: %s\n", assistantMessage.Content)

			// Add assistant response to chat history
			chatHistory = append(chatHistory, assistantMessage)
		} else {
			fmt.Println("No response from the assistant.")
		}
	}
}
