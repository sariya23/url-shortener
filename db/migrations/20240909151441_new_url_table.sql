-- +goose Up
-- +goose StatementBegin
create table if not exists url (
    url_id bigint generated always as identity primary key,
	alias text not null unique,
	url text not null
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists url;
-- +goose StatementEnd
