-- +goose Up
-- +goose StatementBegin
INSERT INTO settings (
    name, value, user_id
) VALUES (
    'torrents_per_day', 10, NULL
), (
    'channel', -1002184825487, NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM settings 
WHERE name = 'torrents_per_day' AND value = 10 AND user_id IS NULL;

DELETE FROM settings 
WHERE name = 'channel' AND value = -1002184825487 AND user_id IS NULL;
-- +goose StatementEnd
