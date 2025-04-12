CREATE TABLE IF NOT EXISTS products
(
    id          UUID PRIMARY KEY,
    dateTime    TIMESTAMP NOT NULL,
    type        TEXT      NOT NULL CHECK (type IN ('электроника', 'одежда', 'обувь')),
    receptionId UUID      NOT NULL REFERENCES receptions (id)
);