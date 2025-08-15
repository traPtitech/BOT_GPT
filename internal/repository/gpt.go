package repository

import (
	"encoding/json"
	"fmt"
	"github.com/openai/openai-go"
)

func SaveMessage(channelID string, index int, message openai.ChatCompletionMessageParamUnion) error {
	converter := &MessageConverter{}
	storedMsg, err := converter.ToStoredMessage(message)
	if err != nil {
		return fmt.Errorf("failed to convert message: %w", err)
	}

	// Serialize StoredMessage to JSON
	msgJSON, err := json.Marshal(storedMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal stored message: %w", err)
	}

	_, err = db.Exec(`INSERT INTO channel_messages (channel_id, message_index, message, message_json, message_role) 
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE message = VALUES(message), message_json = VALUES(message_json), message_role = VALUES(message_role)`,
		channelID, index, msgJSON, msgJSON, storedMsg.Role)

	return err
}

func LoadMessages(channelID string) ([]openai.ChatCompletionMessageParamUnion, error) {
	var messages []openai.ChatCompletionMessageParamUnion
	converter := &MessageConverter{}

	// Try to load from new JSON column first, fallback to old BLOB column
	rows, err := db.Queryx(`SELECT 
		COALESCE(message_json, message) as msg_data,
		message_json IS NOT NULL as is_json
		FROM channel_messages 
		WHERE channel_id = ? 
		ORDER BY message_index`, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var msgData []byte
		var isJSON bool

		err = rows.Scan(&msgData, &isJSON)
		if err != nil {
			return nil, err
		}

		if isJSON {
			// New JSON format
			var storedMsg StoredMessage
			if err := json.Unmarshal(msgData, &storedMsg); err != nil {
				return nil, fmt.Errorf("failed to unmarshal JSON message: %w", err)
			}

			msg, err := converter.FromStoredMessage(&storedMsg)
			if err != nil {
				return nil, fmt.Errorf("failed to convert stored message: %w", err)
			}
			messages = append(messages, msg)
		} else {
			// Old BLOB format - skip for now or attempt to migrate
			// For now, we'll skip old messages to avoid compatibility issues
			continue
		}
	}

	return messages, nil
}

func DeleteMessages(channelID string) error {
	_, err := db.Exec("DELETE FROM channel_messages WHERE channel_id = ?", channelID)
	if err != nil {
		return err
	}

	return nil
}

func GetChannelIDs() ([]string, error) {
	var channelIDs []string
	err := db.Select(&channelIDs, "SELECT DISTINCT channel_id FROM channel_messages")
	if err != nil {
		return nil, err
	}

	return channelIDs, nil
}
