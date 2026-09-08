CREATE TABLE feature_change_links (
    id INTEGER PRIMARY KEY,
    change_item_id INTEGER NOT NULL REFERENCES roadmap_items(id),
    capability_item_id INTEGER NOT NULL REFERENCES roadmap_items(id),
    created_at TEXT NOT NULL,
    CHECK(change_item_id <> capability_item_id),
    UNIQUE(change_item_id, capability_item_id)
);

CREATE INDEX feature_change_links_change_idx ON feature_change_links(change_item_id);
CREATE INDEX feature_change_links_capability_idx ON feature_change_links(capability_item_id);

CREATE TABLE capability_state_history (
    id INTEGER PRIMARY KEY,
    capability_item_id INTEGER NOT NULL REFERENCES roadmap_items(id),
    source_change_item_id INTEGER REFERENCES roadmap_items(id),
    state TEXT NOT NULL,
    accepted_by TEXT NOT NULL,
    accepted_at TEXT NOT NULL,
    created_at TEXT NOT NULL
);

CREATE INDEX capability_state_history_capability_idx ON capability_state_history(capability_item_id, id);

CREATE TABLE dossier_artifacts (
    id INTEGER PRIMARY KEY,
    roadmap_item_id INTEGER NOT NULL REFERENCES roadmap_items(id),
    kind TEXT NOT NULL,
    author_role TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('Draft', 'Accepted', 'Rejected', 'Superseded')) DEFAULT 'Draft',
    summary TEXT NOT NULL,
    evidence TEXT NOT NULL DEFAULT '',
    supersedes_id INTEGER REFERENCES dossier_artifacts(id),
    created_at TEXT NOT NULL,
    accepted_at TEXT,
    accepted_by TEXT
);

CREATE INDEX dossier_artifacts_item_idx ON dossier_artifacts(roadmap_item_id, id);

INSERT INTO capability_state_history(capability_item_id, state, accepted_by, accepted_at, created_at)
SELECT id, current_state, 'migration', updated_at, updated_at
FROM roadmap_items
WHERE feature_type = 'Capability' AND trim(current_state) <> '';
