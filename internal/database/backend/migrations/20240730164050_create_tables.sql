-- +goose Up
-- +goose StatementBegin
CREATE TABLE settings
(
  id      BIGINT NOT NULL GENERATED ALWAYS AS IDENTITY,
  name    TEXT   NOT NULL,
  value   BIGINT NOT NULL,
  user_id BIGINT DEFAULT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE torrent_x_user
(
  torrent_id BIGINT  NOT NULL,
  user_id    BIGINT  NOT NULL,
  sent       BOOLEAN NOT NULL DEFAULT FALSE,
  PRIMARY KEY (torrent_id, user_id)
);

CREATE TABLE torrents
(
  id            BIGINT    NOT NULL GENERATED ALWAYS AS IDENTITY,
  message_id    BIGINT    DEFAULT NULL,
  torrent_link  TEXT      NOT NULL UNIQUE,
  name          TEXT      DEFAULT NULL,
  size          BIGINT    DEFAULT NULL,
  time_added    TIMESTAMP NOT NULL,
  time_started  TIMESTAMP DEFAULT NULL,
  time_finished TIMESTAMP DEFAULT NULL,
  error         TEXT      DEFAULT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE users
(
  id        BIGINT NOT NULL,
  chat_id   BIGINT NOT NULL,
  priority  INT    NOT NULL DEFAULT 0,
  PRIMARY KEY (id)
);

ALTER TABLE torrent_x_user
  ADD CONSTRAINT FK_torrents_TO_torrent_x_user
    FOREIGN KEY (torrent_id)
    REFERENCES torrents (id);

ALTER TABLE torrent_x_user
  ADD CONSTRAINT FK_users_TO_torrent_x_user
    FOREIGN KEY (user_id)
    REFERENCES users (id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE settings;

DROP TABLE torrent_x_user;

DROP TABLE torrents;

DROP TABLE users;
-- +goose StatementEnd
