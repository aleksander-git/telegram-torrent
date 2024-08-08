-- name: AddUser :exec
INSERT INTO users (
    id, chat_id, priority
) VALUES (
    $1, $2, $3
);

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1;

-- name: UpdateUserPriority :exec
UPDATE users 
    SET priority = $2
WHERE id = $1;

-- name: AddTorrent :exec
INSERT INTO torrents (
    torrent_link, time_added
) VALUES (
    $1, $2
);

-- name: GetTorrent :one
SELECT *
FROM torrents
WHERE torrent_link = $1;

-- name: GetFirstUnstartedTorrent :one
SELECT t.*
FROM torrents AS t
INNER JOIN torrent_x_user AS txu
    ON t.id = txu.torrent_id
INNER JOIN users AS u
    ON t.user_id = u.id
WHERE time_started = NULL
ORDER BY time_added ASC, u.priority DESC
LIMIT 1;

-- name: GetUserTorrents :many
SELECT 
    t.*
FROM torrent_x_user AS txu
INNER JOIN torrents AS t
    ON txu.torrent_id = t.id
WHERE txu.user_id = $1;

-- name: UpdateTorrentMessageID :exec
UPDATE torrents
    SET message_id = $2
WHERE torrent_link = $1;

-- name: UpdateTorrentName :exec
UPDATE torrents
    SET name = $2
WHERE torrent_link = $1;

-- name: UpdateTorrentSize :exec
UPDATE torrents
    SET size = $2
WHERE torrent_link = $1;

-- name: UpdateTorrentStatus :exec
UPDATE torrents
    SET time_started = $2, 
    time_finished = $3, 
    error = $4
WHERE torrent_link = $1;

-- name: AddTorrentXUser :exec
INSERT INTO torrent_x_user (
    torrent_id, user_id
) VALUES (
    $1, $2
);

-- name: UpdateTorrentXUser :exec
UPDATE torrent_x_user
    SET sent = $3
WHERE torrent_id = $1 AND user_id = $2;

-- name: GetUnsentUsersForTorrent :many
SELECT 
    u.*
FROM torrents AS t
INNER JOIN torrent_x_user AS txu
    ON t.id = txu.torrent_id
INNER JOIN users AS u
    ON t.user_id = u.id
WHERE t.torrent_link = $1 AND u.sent = FALSE;
