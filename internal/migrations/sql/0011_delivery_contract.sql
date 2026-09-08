ALTER TABLE plan_revisions ADD COLUMN active INTEGER NOT NULL DEFAULT 0 CHECK(active IN (0,1));
ALTER TABLE tasks ADD COLUMN verification TEXT NOT NULL DEFAULT '[]';
ALTER TABLE acceptance_criteria ADD COLUMN plan_revision_id INTEGER REFERENCES plan_revisions(id);
ALTER TABLE acceptance_criteria ADD COLUMN criterion_key TEXT NOT NULL DEFAULT '';
ALTER TABLE acceptance_criteria ADD COLUMN required INTEGER NOT NULL DEFAULT 1 CHECK(required IN (0,1));
ALTER TABLE acceptance_criteria ADD COLUMN waived_by TEXT NOT NULL DEFAULT '';
ALTER TABLE acceptance_criteria ADD COLUMN waiver_reason TEXT NOT NULL DEFAULT '';

CREATE TABLE criterion_evidence (
    id INTEGER PRIMARY KEY,
    acceptance_criterion_id INTEGER NOT NULL REFERENCES acceptance_criteria(id),
    status TEXT NOT NULL CHECK(status IN ('Passed','Failed','Waived')),
    verification_method TEXT NOT NULL,
    evidence TEXT NOT NULL,
    recorded_by TEXT NOT NULL DEFAULT '',
    waiver_reason TEXT NOT NULL DEFAULT '',
    recorded_at TEXT NOT NULL
);

CREATE TABLE task_events (
    id INTEGER PRIMARY KEY,
    task_id INTEGER NOT NULL REFERENCES tasks(id),
    status TEXT NOT NULL CHECK(status IN ('In Progress','Done','Blocked')),
    outcome TEXT NOT NULL,
    recorded_at TEXT NOT NULL
);

CREATE UNIQUE INDEX plan_revisions_active_item_idx ON plan_revisions(roadmap_item_id) WHERE active=1;
CREATE INDEX acceptance_criteria_plan_idx ON acceptance_criteria(plan_revision_id, id);
CREATE INDEX criterion_evidence_criterion_idx ON criterion_evidence(acceptance_criterion_id, id);
CREATE INDEX task_events_task_idx ON task_events(task_id, id);
