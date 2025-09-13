-- +goose Up
CREATE TABLE channel_model_config
(
    channel_id VARCHAR(255) NOT NULL PRIMARY KEY,
    model_name VARCHAR(100) NOT NULL DEFAULT 'gpt-4.1-mini'
);

-- +goose Down
DROP TABLE channel_model_config;
