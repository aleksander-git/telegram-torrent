// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package backend

import (
	"database/sql"
	"time"
)

type Torrent struct {
	ID           int64
	MessageID    sql.NullInt64
	TorrentLink  string
	Name         sql.NullString
	Size         sql.NullInt64
	TimeAdded    time.Time
	TimeStarted  sql.NullTime
	TimeFinished sql.NullTime
	Error        sql.NullString
}

type TorrentXUser struct {
	TorrentID int64
	UserID    int64
	Sent      bool
}

type User struct {
	ID       int64
	ChatID   int64
	Priority int32
}