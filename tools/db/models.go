// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package db

import (
	"database/sql"
)

type GooseDbVersion struct {
	ID        uint64
	VersionID int64
	IsApplied bool
	Tstamp    sql.NullTime
}

type User struct {
	ID        int32
	Username  string
	Password  string
	CreatedAt sql.NullTime
}