package repository

import (
	"encoding/json"
	"fmt"
	"github.com/openai/openai-go"
)

// MessageRole represents the role of a message
type MessageRole string

const (
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleSystem    MessageRole = "system"
	RoleUnknown   MessageRole = "unknown"
)

// RoleHelper provides utilities for working with message roles
type RoleHelper struct{}

// GetMessageRole extracts the role from a ChatCompletionMessageParamUnion
func (rh *RoleHelper) GetMessageRole(msg openai.ChatCompletionMessageParamUnion) MessageRole {
	// Since we can't directly access the role field, we'll serialize and deserialize
	// to extract the role information
	data, err := json.Marshal(msg)
	if err != nil {
		return RoleUnknown
	}
	
	var temp map[string]interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return RoleUnknown
	}
	
	if role, ok := temp["role"].(string); ok {
		switch role {
		case "user":
			return RoleUser
		case "assistant":
			return RoleAssistant
		case "system":
			return RoleSystem
		default:
			return RoleUnknown
		}
	}
	
	return RoleUnknown
}

// HasSystemMessage checks if the message list contains a system message
func (rh *RoleHelper) HasSystemMessage(messages []openai.ChatCompletionMessageParamUnion) bool {
	for _, msg := range messages {
		if rh.GetMessageRole(msg) == RoleSystem {
			return true
		}
	}
	return false
}

// FindSystemMessageIndex returns the index of the first system message, or -1 if not found
func (rh *RoleHelper) FindSystemMessageIndex(messages []openai.ChatCompletionMessageParamUnion) int {
	for i, msg := range messages {
		if rh.GetMessageRole(msg) == RoleSystem {
			return i
		}
	}
	return -1
}

// GetLastMessageRole returns the role of the last message in the list
func (rh *RoleHelper) GetLastMessageRole(messages []openai.ChatCompletionMessageParamUnion) MessageRole {
	if len(messages) == 0 {
		return RoleUnknown
	}
	return rh.GetMessageRole(messages[len(messages)-1])
}

// CountMessagesByRole counts messages by their role
func (rh *RoleHelper) CountMessagesByRole(messages []openai.ChatCompletionMessageParamUnion) map[MessageRole]int {
	counts := make(map[MessageRole]int)
	for _, msg := range messages {
		role := rh.GetMessageRole(msg)
		counts[role]++
	}
	return counts
}

// ValidateMessageSequence checks if the message sequence follows proper conversational patterns
func (rh *RoleHelper) ValidateMessageSequence(messages []openai.ChatCompletionMessageParamUnion) error {
	if len(messages) == 0 {
		return nil
	}
	
	// Check if system message is first (if present)
	systemIndex := rh.FindSystemMessageIndex(messages)
	if systemIndex != -1 && systemIndex != 0 {
		return fmt.Errorf("system message should be the first message, found at index %d", systemIndex)
	}
	
	// Check for proper alternation after system message (optional validation)
	startIndex := 0
	if systemIndex == 0 {
		startIndex = 1
	}
	
	for i := startIndex; i < len(messages); i++ {
		role := rh.GetMessageRole(messages[i])
		if role == RoleSystem {
			return fmt.Errorf("system message found at unexpected position %d", i)
		}
		
		// Note: We're allowing flexible conversation flow
		// Multiple user messages or assistant messages in sequence are acceptable
	}
	
	return nil
}

// GetMessageContent attempts to extract text content from a message
func (rh *RoleHelper) GetMessageContent(msg openai.ChatCompletionMessageParamUnion) string {
	data, err := json.Marshal(msg)
	if err != nil {
		return ""
	}
	
	var temp map[string]interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return ""
	}
	
	// Try to extract content as string
	if content, ok := temp["content"].(string); ok {
		return content
	}
	
	// Try to extract content from array (for multi-part messages)
	if contentArray, ok := temp["content"].([]interface{}); ok {
		for _, part := range contentArray {
			if partMap, ok := part.(map[string]interface{}); ok {
				if text, ok := partMap["text"].(string); ok {
					return text // Return first text part
				}
			}
		}
	}
	
	return ""
}