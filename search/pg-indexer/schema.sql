CREATE TABLE IF NOT EXISTS entities
(
    id             TEXT PRIMARY KEY,
    last_updated   TIMESTAMP NOT NULL,

    type           TEXT      NOT NULL,
    addr           TEXT      NOT NULL,
    version        TEXT      NOT NULL,
    title          TEXT      NOT NULL,
    description    TEXT,
    link_variables JSONB
);

CREATE TABLE IF NOT EXISTS import_jobs
(
    id           SERIAL PRIMARY KEY,

    created_at   TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    successful   BOOLEAN
);