CREATE TABLE torrents
(
  user_id        BIGINT    NOT NULL,
  torrent_link   TEXT      NOT NULL,
  telegram_link  TEXT      NULL     DEFAULT NULL,
  time_added     TIMESTAMP NOT NULL,
  time_started   TIMESTAMP NULL     DEFAULT NULL,
  time_finished  TIMESTAMP NULL     DEFAULT NULL,
  error          TEXT      NULL     DEFAULT NULL
);

CREATE TABLE users
(
  id            BIGINT NOT NULL,
  telegram_name TEXT   NOT NULL,
  priority      INT    NOT NULL DEFAULT 0,
  PRIMARY KEY (id)
);

ALTER TABLE torrents
  ADD CONSTRAINT FK_users_TO_torrents
    FOREIGN KEY (user_id)
    REFERENCES users (id);
