package repository

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/sashabaranov/go-openai"
)

type Repository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

var db *sqlx.DB

func InitDB(database *sqlx.DB) {
	db = database
}

// GetModelForChannel チャンネルのモデル設定を取得
func GetModelForChannel(channelID string) (string, error) {
	var model string
	err := db.Get(&model, "SELECT model_name FROM channel_model_config WHERE channel_id = ?", channelID)
	if err != nil {
		if err == sql.ErrNoRows {
			// レコードが存在しない場合はデフォルトモデルを返す
			return openai.GPT4o, nil
		}

		return "", err
	}

	return model, nil
}

// SetModelForChannel チャンネルのモデル設定を保存
func SetModelForChannel(channelID, model string) error {
	_, err := db.Exec("INSERT INTO channel_model_config (channel_id, model_name) VALUES (?, ?) ON DUPLICATE KEY UPDATE model_name = VALUES(model_name)", channelID, model)

	return err
}
