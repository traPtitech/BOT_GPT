package repository

import (
	"github.com/jmoiron/sqlx"
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
