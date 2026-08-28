PRAGMA defer_foreign_keys = ON;
CREATE TABLE roadmap_items_next (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    category_id INTEGER NOT NULL REFERENCES categories(id),
    horizon TEXT NOT NULL CHECK(horizon IN ('Now', 'Next', 'Later')),
    status TEXT NOT NULL CHECK(status IN ('Planned', 'Ready', 'In Progress', 'Blocked', 'Done', 'Won’t Do', 'Archived')) DEFAULT 'Planned',
    rationale TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
INSERT INTO roadmap_items_next SELECT id,title,description,category_id,horizon,CASE status WHEN 'Proposed' THEN 'Planned' WHEN 'Approved' THEN 'Ready' WHEN 'Active' THEN 'In Progress' WHEN 'Completed' THEN 'Done' WHEN 'Declined' THEN 'Won’t Do' ELSE status END,rationale,created_at,updated_at FROM roadmap_items;
DROP TABLE roadmap_items;
ALTER TABLE roadmap_items_next RENAME TO roadmap_items;
CREATE INDEX roadmap_items_status_idx ON roadmap_items(status);
CREATE TABLE tasks_next (
    id INTEGER PRIMARY KEY,
    plan_revision_id INTEGER NOT NULL REFERENCES plan_revisions(id),
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL CHECK(status IN ('To Do', 'In Progress', 'Done', 'Blocked')) DEFAULT 'To Do',
    outcome TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    started_at TEXT,
    completed_at TEXT,
    blocked_at TEXT
);
INSERT INTO tasks_next SELECT id,plan_revision_id,title,description,CASE status WHEN 'Pending' THEN 'To Do' WHEN 'Active' THEN 'In Progress' WHEN 'Completed' THEN 'Done' ELSE status END,outcome,created_at,started_at,completed_at,blocked_at FROM tasks;
DROP TABLE tasks;
ALTER TABLE tasks_next RENAME TO tasks;
CREATE INDEX tasks_plan_revision_idx ON tasks(plan_revision_id);
