package main // This file belongs to the 'handlers' package

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// Global variables

// Holding the entire conversation
var conversationHistory []OpenRouterMessage

// A global map to hold our dummy learning materials.
var learningMaterials = map[string][]struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}{
	"articles": {
		{Title: "World War II | Wikipedia", URL: "https://en.wikipedia.org/wiki/World_War_II"},
		{Title: "Machine Learning | Wikipedia ", URL: "https://en.wikipedia.org/wiki/Machine_learning"},
		{Title: "Gravity | Wikipedia", URL: "https://en.wikipedia.org/wiki/Gravity"},
		{Title: "Computer Networking | Wikipedia", URL: "https://en.wikipedia.org/wiki/Computer_network"},
		{Title: "Algebraic Equation | Wikipedia", URL: "https://en.wikipedia.org/wiki/Algebraic_equation"},
	},
	"videos": {
		{Title: "World War 2 Part 1 | Oversimplified", URL: "https://www.youtube.com/watch?v=_uk_6vfqwTA"},
		{Title: "BIOLOGY explained in 17 Minutes", URL: "https://www.youtube.com/watch?v=3tisOnOkwzo"},
		{Title: "3Blue1Brown: Essence of Calculus", URL: "https://www.youtube.com/watch?v=WUvTyaaNkzM"},
		{Title: "Periodic Tables | Khan Academy", URL: "https://www.youtube.com/watch?v=Jg3Fo4WD8XE"},
		{Title: "3Blue1Brown: Neural Networks", URL: "https://www.youtube.com/watch?v=aircAruvnKk"},
	},
}

// A global slice of keywords to check for learning materials requests.
var learningKeywords = []string{"materi", "materials", "video", "article", "artikel", "videos", "tambahan materi", "materi externl"}

// Structs to match the OpenRouter API payload
type OpenRouterImageURL struct {
	URL string `json:"url"`
}

type OpenRouterContent struct {
	Type     string              `json:"type"`
	Text     string              `json:"text,omitempty"`
	ImageURL *OpenRouterImageURL `json:"image_url,omitempty"`
}

type OpenRouterMessage struct {
	Role    string              `json:"role"`
	Content []OpenRouterContent `json:"content"`
}

type OpenRouterPayload struct {
	Model    string              `json:"model"`
	Messages []OpenRouterMessage `json:"messages"`
}

// Helper method to hardcode check if the user's message asks for external materials
// Gives dummy data to be shown to the frontend
func handleLearningMaterialsRequest(c *gin.Context, message string) bool {
	lowerMessage := strings.ToLower(message)

	// Loop through our list of keywords
	for _, keyword := range learningKeywords {
		if strings.Contains(lowerMessage, keyword) {
			// If we find a keyword, send the custom JSON response
			c.JSON(http.StatusOK, gin.H{
				"bot_response": "Of course! Here are some learning materials I found for you:",
				"articles":     learningMaterials["articles"],
				"videos":       learningMaterials["videos"],
			})
			return true
		}
	}
	return false
}

// HandleChat processes the user's message and image.
// Using OpenRouter to connect to
func HandleChatOPENROUTER(c *gin.Context) {
	// Get the user's message from the form data
	message := c.PostForm("message")

	// Get the uploaded image file if one exists
	fileHeader, err := c.FormFile("image")
	var imagePart *OpenRouterContent

	// Process the image into a Base64 string if one was uploaded
	if err == nil && fileHeader != nil {
		file, err := fileHeader.Open()
		if err != nil {
			log.Printf("Failed to open uploaded file: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
			return
		}
		defer file.Close()

		fileBytes, err := io.ReadAll(file)
		if err != nil {
			log.Printf("Failed to read uploaded file: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
			return
		}

		base64Image := base64.StdEncoding.EncodeToString(fileBytes)
		dataURL := fmt.Sprintf("data:%s;base64,%s", fileHeader.Header.Get("Content-Type"), base64Image)

		imagePart = &OpenRouterContent{
			Type:     "image_url",
			ImageURL: &OpenRouterImageURL{URL: dataURL},
		}
		log.Printf("Successfully processed image with MIME type: %s", fileHeader.Header.Get("Content-Type"))
	}

	// Assemble the content parts (text and/or image) for the user's message
	var messageParts []OpenRouterContent
	if message != "" {
		messageParts = append(messageParts, OpenRouterContent{
			Type: "text",
			Text: message,
		})
	}
	if imagePart != nil {
		messageParts = append(messageParts, *imagePart)
	}

	// Add the new user message to our conversation history
	conversationHistory = append(conversationHistory, OpenRouterMessage{
		Role:    "user",
		Content: messageParts,
	})

	// Check if the user is asking for learning materials and handle the response
	// calling the helper method
	if handleLearningMaterialsRequest(c, message) {
		return // EXIT FROM THE PROCESS the helper handled the request
	}

	var fullMessages []OpenRouterMessage
	// Add a system message to guide the bot's tone and formatting
	fullMessages = append(fullMessages, OpenRouterMessage{
		Role: "system",
		Content: []OpenRouterContent{
			{Type: "text", Text: "You are a helpful learning assistant. Please answer all questions in plain text without using any Markdown formatting, such as bold, italics, or lists. Your responses should be clean and simple."},
		},
	})

	// Now add the entire conversation history after the system message
	fullMessages = append(fullMessages, conversationHistory...) // The '...' unpacks the slice

	// Building the final payload as a struct
	payload := OpenRouterPayload{
		Model:    "google/gemini-flash-1.5",
		Messages: fullMessages,
	}

	fmt.Printf("%+v\n", payload)

	// Marshal the struct into a JSON payload
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create API payload."})
		return
	}

	// Build the HTTP request
	url := "https://openrouter.ai/api/v1/chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create HTTP request."})
		return
	}

	// Send the request using a new HTTP client
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENROUTER_API_KEY"))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("HTTP request error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send API request."})
		return
	}
	defer resp.Body.Close() // Making sure no resource leaks

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read API response."})
		return
	}

	// Unmarshal the JSON response
	var apiResponse map[string]interface{}
	json.Unmarshal(body, &apiResponse)

	// Extract the bot's response using safe type assertions
	var botResponse string
	if choices, ok := apiResponse["choices"].([]interface{}); ok && len(choices) > 0 {
		if firstChoice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := firstChoice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					botResponse = content
				}
			}
		}
	} else {
		botResponse = "I'm sorry, I couldn't generate a response."
	}

	// Add the chatbot's response to the conversation history
	if botResponse != "" {
		conversationHistory = append(conversationHistory, OpenRouterMessage{
			Role:    "assistant",
			Content: []OpenRouterContent{{Type: "text", Text: botResponse}},
		})
	}

	// Send the response back to the frontend as JSON!
	c.JSON(http.StatusOK, gin.H{"bot_response": botResponse})
}
