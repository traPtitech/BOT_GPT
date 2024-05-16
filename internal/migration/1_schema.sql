-- +goose Up
CREATE TABLE channel_messages
(
    channel_id    VARCHAR(255) NOT NULL,
    message_index INT          NOT NULL,
    message       BLOB         NOT NULL,
    PRIMARY KEY (channel_id, message_index)
);
