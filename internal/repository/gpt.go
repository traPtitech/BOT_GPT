package repository

import (
	"bytes"
	"encoding/gob"

	"github.com/openai/openai-go/v2"
)

func SaveMessage(channelID string, index int, message openai.ChatCompletionMessage) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(message)
	if err != nil {
		return err
	}

	_, err = db.Exec(`INSERT INTO channel_messages (channel_id, message_index, message) VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE message = VALUES(message)`,
		channelID, index, buf.Bytes())

	return err
}

func LoadMessages(channelID string) ([]openai.ChatCompletionMessage, error) {
	var messages []openai.ChatCompletionMessage
	rows, err := db.Queryx("SELECT message FROM channel_messages WHERE channel_id = ? ORDER BY message_index", channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var binaryMessage []byte
		var message openai.ChatCompletionMessage

		err = rows.Scan(&binaryMessage)
		if err != nil {
			return nil, err
		}

		buf := bytes.NewBuffer(binaryMessage)
		dec := gob.NewDecoder(buf)
		err = dec.Decode(&message)
		if err != nil {
			return nil, err
		}

		messages = append(messages, message)
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
