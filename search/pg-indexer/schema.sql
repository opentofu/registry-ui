-- This file is to be used as reference on how to create the tables, this is not automatically run anywhere
-- Please attempt to keep this file up to date with the actual schema
CREATE TABLE IF NOT EXISTS entities
(
    id             TEXT PRIMARY KEY,
    last_updated   TIMESTAMP NOT NULL,

    type           TEXT      NOT NULL,
    addr           TEXT      NOT NULL,
    version        TEXT      NOT NULL,
    title          TEXT      NOT NULL,
    description    TEXT,
    link_variables JSONB,
    document       TSVECTOR
);

CREATE INDEX IF NOT EXISTS idx_entities_title_lower
    ON entities (lower(title));

CREATE TABLE IF NOT EXISTS import_jobs
(
    id           SERIAL PRIMARY KEY,

    created_at   TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    successful   BOOLEAN
);

ALTER TABLE entities ADD COLUMN popularity INT DEFAULT 0;
