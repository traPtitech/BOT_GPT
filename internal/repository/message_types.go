package repository

import (
	"encoding/json"
	"fmt"
	"github.com/openai/openai-go"
)

// StoredMessage represents a message as stored in the database
type StoredMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
	Type    string          `json:"type"` // "text", "image", "multi"
}

// SimpleMessage is a simplified representation for JSON storage
type SimpleMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
	Type    string      `json:"type"`
}

// MessageConverter handles conversion between SDK types and storage types
type MessageConverter struct{}

// ToStoredMessage converts an OpenAI message to a storable format
func (mc *MessageConverter) ToStoredMessage(msg openai.ChatCompletionMessageParamUnion) (*StoredMessage, error) {
	// Serialize the entire message as JSON for now
	// This is a simpler approach that avoids complex type assertions
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}
	
	// Try to extract role from the JSON for indexing
	var temp map[string]interface{}
	role := "unknown"
	if json.Unmarshal(data, &temp) == nil {
		if r, ok := temp["role"].(string); ok {
			role = r
		}
	}
	
	return &StoredMessage{
		Role:    role,
		Content: data,
		Type:    "json",
	}, nil
}

// FromStoredMessage converts a stored message back to an OpenAI message
func (mc *MessageConverter) FromStoredMessage(stored *StoredMessage) (openai.ChatCompletionMessageParamUnion, error) {
	// First try to unmarshal as a simple message structure
	var simple SimpleMessage
	if err := json.Unmarshal(stored.Content, &simple); err != nil {
		return openai.UserMessage(""), fmt.Errorf("failed to unmarshal stored message: %w", err)
	}
	
	// Convert based on role
	switch simple.Role {
	case "user":
		return mc.restoreUserMessage(&simple)
	case "assistant":
		return mc.restoreAssistantMessage(&simple)
	case "system":
		return mc.restoreSystemMessage(&simple)
	default:
		// Fallback to user message with content as string
		contentStr := fmt.Sprintf("%v", simple.Content)
		return openai.UserMessage(contentStr), nil
	}
}

func (mc *MessageConverter) restoreUserMessage(simple *SimpleMessage) (openai.ChatCompletionMessageParamUnion, error) {
	// Handle different content types
	switch content := simple.Content.(type) {
	case string:
		return openai.UserMessage(content), nil
	case []interface{}:
		// Try to restore as multi-content message
		// For now, just extract text content as fallback
		for _, part := range content {
			if partMap, ok := part.(map[string]interface{}); ok {
				if text, ok := partMap["text"].(string); ok {
					return openai.UserMessage(text), nil
				}
			}
		}
		return openai.UserMessage(""), nil
	default:
		contentStr := fmt.Sprintf("%v", content)
		return openai.UserMessage(contentStr), nil
	}
}

func (mc *MessageConverter) restoreAssistantMessage(simple *SimpleMessage) (openai.ChatCompletionMessageParamUnion, error) {
	contentStr := fmt.Sprintf("%v", simple.Content)
	return openai.AssistantMessage(contentStr), nil
}

func (mc *MessageConverter) restoreSystemMessage(simple *SimpleMessage) (openai.ChatCompletionMessageParamUnion, error) {
	contentStr := fmt.Sprintf("%v", simple.Content)
	return openai.SystemMessage(contentStr), nil
}