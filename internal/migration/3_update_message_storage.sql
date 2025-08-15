-- +goose Up
-- Update message storage to use JSON format for new SDK compatibility
ALTER TABLE channel_messages 
ADD COLUMN message_json JSON NULL AFTER message,
ADD COLUMN message_role VARCHAR(20) NULL AFTER message_json,
ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- +goose Down
ALTER TABLE channel_messages 
DROP COLUMN message_json,
DROP COLUMN message_role,
DROP COLUMN created_at;