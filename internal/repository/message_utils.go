package repository

import (
	"fmt"
	"github.com/openai/openai-go"
	"time"
)

// MessageUtils provides utility functions for working with messages
type MessageUtils struct {
	roleHelper *RoleHelper
}

// NewMessageUtils creates a new MessageUtils instance
func NewMessageUtils() *MessageUtils {
	return &MessageUtils{
		roleHelper: &RoleHelper{},
	}
}

// AddMessageWithValidation adds a message to the channel with role validation
func (mu *MessageUtils) AddMessageWithValidation(
	channelID string, 
	messages *[]openai.ChatCompletionMessageParamUnion, 
	newMessage openai.ChatCompletionMessageParamUnion,
) error {
	// Create a temporary slice to test validation
	testMessages := append(*messages, newMessage)
	
	// Validate the sequence
	if err := mu.roleHelper.ValidateMessageSequence(testMessages); err != nil {
		return fmt.Errorf("message sequence validation failed: %w", err)
	}
	
	// If validation passes, add the message
	*messages = testMessages
	
	// Save to database
	index := len(*messages) - 1
	if err := SaveMessage(channelID, index, newMessage); err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}
	
	return nil
}

// CleanupOldMessages removes messages older than the specified duration
func (mu *MessageUtils) CleanupOldMessages(channelID string, maxAge time.Duration) error {
	// This would require adding timestamps to messages
	// For now, we'll implement a basic cleanup based on message count
	return mu.TrimMessageHistory(channelID, 100) // Keep last 100 messages
}

// TrimMessageHistory keeps only the most recent N messages
func (mu *MessageUtils) TrimMessageHistory(channelID string, maxMessages int) error {
	messages, err := LoadMessages(channelID)
	if err != nil {
		return fmt.Errorf("failed to load messages: %w", err)
	}
	
	if len(messages) <= maxMessages {
		return nil // No trimming needed
	}
	
	// Keep the system message (if it exists) and the most recent messages
	systemIndex := mu.roleHelper.FindSystemMessageIndex(messages)
	var trimmedMessages []openai.ChatCompletionMessageParamUnion
	
	if systemIndex == 0 {
		// Keep system message + most recent user/assistant messages
		recentMessages := messages[max(1, len(messages)-maxMessages+1):]
		trimmedMessages = append([]openai.ChatCompletionMessageParamUnion{messages[0]}, recentMessages...)
	} else {
		// No system message, just keep recent messages
		trimmedMessages = messages[max(0, len(messages)-maxMessages):]
	}
	
	// Delete all messages for this channel
	if err := DeleteMessages(channelID); err != nil {
		return fmt.Errorf("failed to delete old messages: %w", err)
	}
	
	// Re-save the trimmed messages
	for i, msg := range trimmedMessages {
		if err := SaveMessage(channelID, i, msg); err != nil {
			return fmt.Errorf("failed to save trimmed message %d: %w", i, err)
		}
	}
	
	return nil
}

// GetConversationSummary returns a summary of the conversation
func (mu *MessageUtils) GetConversationSummary(messages []openai.ChatCompletionMessageParamUnion) map[string]interface{} {
	roleCounts := mu.roleHelper.CountMessagesByRole(messages)
	
	summary := map[string]interface{}{
		"total_messages":     len(messages),
		"user_messages":      roleCounts[RoleUser],
		"assistant_messages": roleCounts[RoleAssistant],
		"system_messages":    roleCounts[RoleSystem],
		"has_system":         mu.roleHelper.HasSystemMessage(messages),
		"last_role":          string(mu.roleHelper.GetLastMessageRole(messages)),
	}
	
	if len(messages) > 0 {
		firstContent := mu.roleHelper.GetMessageContent(messages[0])
		lastContent := mu.roleHelper.GetMessageContent(messages[len(messages)-1])
		
		summary["first_message_preview"] = truncateString(firstContent, 50)
		summary["last_message_preview"] = truncateString(lastContent, 50)
	}
	
	return summary
}

// RebuildMessageIndices rebuilds the message indices in the database
func (mu *MessageUtils) RebuildMessageIndices(channelID string, messages []openai.ChatCompletionMessageParamUnion) error {
	// Delete all existing messages
	if err := DeleteMessages(channelID); err != nil {
		return fmt.Errorf("failed to delete existing messages: %w", err)
	}
	
	// Re-save all messages with correct indices
	for i, msg := range messages {
		if err := SaveMessage(channelID, i, msg); err != nil {
			return fmt.Errorf("failed to save message %d: %w", i, err)
		}
	}
	
	return nil
}

// Helper functions

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}