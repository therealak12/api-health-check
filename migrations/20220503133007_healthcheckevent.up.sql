CREATE TABLE IF NOT EXISTS healthcheck_events(
    id bigserial PRIMARY KEY,
    healthcheck_id BIGINT DEFAULT 1,
    status TEXT NOT NULL,
    created_at timestamp NOT NULL DEFAULT now(),
    CONSTRAINT fk_category FOREIGN KEY (healthcheck_id) REFERENCES healthchecks (id) ON DELETE SET DEFAULT
);
