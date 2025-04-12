CREATE TABLE IF NOT EXISTS receptions
(
    id       UUID PRIMARY KEY,
    dateTime TIMESTAMP NOT NULL,
    pvzId    UUID      NOT NULL REFERENCES pvz (id),
    status   TEXT      NOT NULL CHECK (status IN ('in_progress', 'close'))
);