-- +goose Up
-- +goose StatementBegin
CREATE TABLE "group" (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    UNIQUE (name)
);
CREATE TABLE song (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    group_id BIGINT NOT NULL REFERENCES "group",
    release DATE NOT NULL,
    link VARCHAR(256) NOT NULL,
    text TEXT NOT NULL,
    UNIQUE(group_id, name)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE song;
DROP TABLE "group";
-- +goose StatementEnd
