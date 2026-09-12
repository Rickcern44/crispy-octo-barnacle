CREATE TABLE epics (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL CHECK(trim(title) <> ''),
    description TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

ALTER TABLE roadmap_items ADD COLUMN epic_id INTEGER REFERENCES epics(id);

CREATE INDEX roadmap_items_epic_idx ON roadmap_items(epic_id);
