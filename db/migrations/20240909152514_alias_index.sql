-- +goose Up
-- +goose StatementBegin
create unique index if not exists url_idx on url(alias);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop index if exists url_idx;
-- +goose StatementEnd
