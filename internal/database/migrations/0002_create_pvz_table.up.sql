CREATE TABLE IF NOT EXISTS pvz
(
    id               UUID PRIMARY KEY,
    registrationDate TIMESTAMP NOT NULL,
    city             TEXT      NOT NULL CHECK (city IN ('Москва', 'Санкт-Петербург', 'Казань'))
);