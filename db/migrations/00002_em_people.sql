-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS em_people1(
    id SERIAL PRIMARY KEY,
    fname varchar(50) NOT NULL,
    surname varchar(100) NOT NULL,
    patronymic varchar(100) NOT NULL,
    age SMALLINT NOT NULL CHECK (age > 0 AND age < 200),
    nationality varchar(50) NOT NULL,
    gender VARCHAR(10) NOT NULL CHECK (gender IN ('male', 'female'))
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS em_people1;
-- +goose StatementEnd
