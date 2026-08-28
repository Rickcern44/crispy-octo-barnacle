CREATE TABLE roadmap_items (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    category_id INTEGER NOT NULL REFERENCES categories(id),
    horizon TEXT NOT NULL CHECK(horizon IN ('Now', 'Next', 'Later')),
    status TEXT NOT NULL CHECK(status IN ('Proposed', 'Approved', 'Active', 'Blocked', 'Completed', 'Declined', 'Archived')) DEFAULT 'Proposed',
    rationale TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE plan_revisions (
    id INTEGER PRIMARY KEY,
    roadmap_item_id INTEGER NOT NULL REFERENCES roadmap_items(id),
    revision INTEGER NOT NULL,
    content TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('Draft', 'Approved')) DEFAULT 'Draft',
    created_at TEXT NOT NULL,
    approved_at TEXT,
    UNIQUE(roadmap_item_id, revision)
);

CREATE TABLE tasks (
    id INTEGER PRIMARY KEY,
    plan_revision_id INTEGER NOT NULL REFERENCES plan_revisions(id),
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL CHECK(status IN ('Pending', 'Active', 'Completed', 'Blocked')) DEFAULT 'Pending',
    outcome TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    started_at TEXT,
    completed_at TEXT,
    blocked_at TEXT
);

CREATE INDEX roadmap_items_status_idx ON roadmap_items(status);
CREATE INDEX plan_revisions_item_idx ON plan_revisions(roadmap_item_id, revision);
CREATE INDEX tasks_plan_revision_idx ON tasks(plan_revision_id);
