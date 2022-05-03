CREATE TABLE IF NOT EXISTS healthchecks(
    id bigserial PRIMARY KEY,
    interval_seconds INTEGER NOT NULL,
    url TEXT NOT NULL,
    http_method VARCHAR (7) NOT NULL, /* OPTIONS has the max length */
    headers_json TEXT,
    body TEXT
);
