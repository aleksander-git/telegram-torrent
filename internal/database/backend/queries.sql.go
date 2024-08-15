// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: queries.sql

package backend

import (
	"context"
	"database/sql"
	"time"
)

const addTorrent = `-- name: AddTorrent :exec
INSERT INTO torrents (
    torrent_link, time_added
) VALUES (
    $1, $2
)
`

type AddTorrentParams struct {
	TorrentLink string
	TimeAdded   time.Time
}

func (q *Queries) AddTorrent(ctx context.Context, arg AddTorrentParams) error {
	_, err := q.db.ExecContext(ctx, addTorrent, arg.TorrentLink, arg.TimeAdded)
	return err
}

const addTorrentXUser = `-- name: AddTorrentXUser :exec
INSERT INTO torrent_x_user (
    torrent_id, user_id
) VALUES (
    $1, $2
)
`

type AddTorrentXUserParams struct {
	TorrentID int64
	UserID    int64
}

func (q *Queries) AddTorrentXUser(ctx context.Context, arg AddTorrentXUserParams) error {
	_, err := q.db.ExecContext(ctx, addTorrentXUser, arg.TorrentID, arg.UserID)
	return err
}

const addUser = `-- name: AddUser :exec
INSERT INTO users (
    id, chat_id, priority
) VALUES (
    $1, $2, $3
)
`

type AddUserParams struct {
	ID       int64
	ChatID   int64
	Priority int32
}

func (q *Queries) AddUser(ctx context.Context, arg AddUserParams) error {
	_, err := q.db.ExecContext(ctx, addUser, arg.ID, arg.ChatID, arg.Priority)
	return err
}

const getFirstUnstartedTorrent = `-- name: GetFirstUnstartedTorrent :one
SELECT t.id, t.message_id, t.torrent_link, t.name, t.size, t.time_added, t.time_started, t.time_finished, t.error
FROM torrents AS t
INNER JOIN torrent_x_user AS txu
    ON t.id = txu.torrent_id
INNER JOIN users AS u
    ON t.user_id = u.id
WHERE time_started = NULL
ORDER BY time_added ASC, u.priority DESC
LIMIT 1
`

func (q *Queries) GetFirstUnstartedTorrent(ctx context.Context) (Torrent, error) {
	row := q.db.QueryRowContext(ctx, getFirstUnstartedTorrent)
	var i Torrent
	err := row.Scan(
		&i.ID,
		&i.MessageID,
		&i.TorrentLink,
		&i.Name,
		&i.Size,
		&i.TimeAdded,
		&i.TimeStarted,
		&i.TimeFinished,
		&i.Error,
	)
	return i, err
}

const getSetting = `-- name: GetSetting :one
SELECT value
FROM settings
WHERE name = $1 AND (user_id == $2 OR user_id IS NULL)
ORDER BY user_id ASC NULLS LAST
LIMIT 1
`

type GetSettingParams struct {
	Name   string
	UserID sql.NullInt64
}

func (q *Queries) GetSetting(ctx context.Context, arg GetSettingParams) (string, error) {
	row := q.db.QueryRowContext(ctx, getSetting, arg.Name, arg.UserID)
	var value string
	err := row.Scan(&value)
	return value, err
}

const getTorrent = `-- name: GetTorrent :one
SELECT id, message_id, torrent_link, name, size, time_added, time_started, time_finished, error
FROM torrents
WHERE torrent_link = $1
`

func (q *Queries) GetTorrent(ctx context.Context, torrentLink string) (Torrent, error) {
	row := q.db.QueryRowContext(ctx, getTorrent, torrentLink)
	var i Torrent
	err := row.Scan(
		&i.ID,
		&i.MessageID,
		&i.TorrentLink,
		&i.Name,
		&i.Size,
		&i.TimeAdded,
		&i.TimeStarted,
		&i.TimeFinished,
		&i.Error,
	)
	return i, err
}

const getUnsentUsersForTorrent = `-- name: GetUnsentUsersForTorrent :many
SELECT 
    u.id, u.chat_id, u.priority
FROM torrents AS t
INNER JOIN torrent_x_user AS txu
    ON t.id = txu.torrent_id
INNER JOIN users AS u
    ON t.user_id = u.id
WHERE t.torrent_link = $1 AND u.sent = FALSE
`

func (q *Queries) GetUnsentUsersForTorrent(ctx context.Context, torrentLink string) ([]User, error) {
	rows, err := q.db.QueryContext(ctx, getUnsentUsersForTorrent, torrentLink)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []User
	for rows.Next() {
		var i User
		if err := rows.Scan(&i.ID, &i.ChatID, &i.Priority); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getUser = `-- name: GetUser :one
SELECT id, chat_id, priority FROM users
WHERE id = $1
`

func (q *Queries) GetUser(ctx context.Context, id int64) (User, error) {
	row := q.db.QueryRowContext(ctx, getUser, id)
	var i User
	err := row.Scan(&i.ID, &i.ChatID, &i.Priority)
	return i, err
}

const getUserTorrents = `-- name: GetUserTorrents :many
SELECT 
    t.id, t.message_id, t.torrent_link, t.name, t.size, t.time_added, t.time_started, t.time_finished, t.error
FROM torrent_x_user AS txu
INNER JOIN torrents AS t
    ON txu.torrent_id = t.id
WHERE txu.user_id = $1
`

func (q *Queries) GetUserTorrents(ctx context.Context, userID int64) ([]Torrent, error) {
	rows, err := q.db.QueryContext(ctx, getUserTorrents, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Torrent
	for rows.Next() {
		var i Torrent
		if err := rows.Scan(
			&i.ID,
			&i.MessageID,
			&i.TorrentLink,
			&i.Name,
			&i.Size,
			&i.TimeAdded,
			&i.TimeStarted,
			&i.TimeFinished,
			&i.Error,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateTorrentMessageID = `-- name: UpdateTorrentMessageID :exec
UPDATE torrents
    SET message_id = $2
WHERE torrent_link = $1
`

type UpdateTorrentMessageIDParams struct {
	TorrentLink string
	MessageID   sql.NullInt64
}

func (q *Queries) UpdateTorrentMessageID(ctx context.Context, arg UpdateTorrentMessageIDParams) error {
	_, err := q.db.ExecContext(ctx, updateTorrentMessageID, arg.TorrentLink, arg.MessageID)
	return err
}

const updateTorrentName = `-- name: UpdateTorrentName :exec
UPDATE torrents
    SET name = $2
WHERE torrent_link = $1
`

type UpdateTorrentNameParams struct {
	TorrentLink string
	Name        sql.NullString
}

func (q *Queries) UpdateTorrentName(ctx context.Context, arg UpdateTorrentNameParams) error {
	_, err := q.db.ExecContext(ctx, updateTorrentName, arg.TorrentLink, arg.Name)
	return err
}

const updateTorrentSize = `-- name: UpdateTorrentSize :exec
UPDATE torrents
    SET size = $2
WHERE torrent_link = $1
`

type UpdateTorrentSizeParams struct {
	TorrentLink string
	Size        sql.NullInt64
}

func (q *Queries) UpdateTorrentSize(ctx context.Context, arg UpdateTorrentSizeParams) error {
	_, err := q.db.ExecContext(ctx, updateTorrentSize, arg.TorrentLink, arg.Size)
	return err
}

const updateTorrentStatus = `-- name: UpdateTorrentStatus :exec
UPDATE torrents
    SET time_started = $2, 
    time_finished = $3, 
    error = $4
WHERE torrent_link = $1
`

type UpdateTorrentStatusParams struct {
	TorrentLink  string
	TimeStarted  sql.NullTime
	TimeFinished sql.NullTime
	Error        sql.NullString
}

func (q *Queries) UpdateTorrentStatus(ctx context.Context, arg UpdateTorrentStatusParams) error {
	_, err := q.db.ExecContext(ctx, updateTorrentStatus,
		arg.TorrentLink,
		arg.TimeStarted,
		arg.TimeFinished,
		arg.Error,
	)
	return err
}

const updateTorrentXUser = `-- name: UpdateTorrentXUser :exec
UPDATE torrent_x_user
    SET sent = $3
WHERE torrent_id = $1 AND user_id = $2
`

type UpdateTorrentXUserParams struct {
	TorrentID int64
	UserID    int64
	Sent      bool
}

func (q *Queries) UpdateTorrentXUser(ctx context.Context, arg UpdateTorrentXUserParams) error {
	_, err := q.db.ExecContext(ctx, updateTorrentXUser, arg.TorrentID, arg.UserID, arg.Sent)
	return err
}

const updateUserPriority = `-- name: UpdateUserPriority :exec
UPDATE users 
    SET priority = $2
WHERE id = $1
`

type UpdateUserPriorityParams struct {
	ID       int64
	Priority int32
}

func (q *Queries) UpdateUserPriority(ctx context.Context, arg UpdateUserPriorityParams) error {
	_, err := q.db.ExecContext(ctx, updateUserPriority, arg.ID, arg.Priority)
	return err
}
