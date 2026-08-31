CREATE TABLE feature_phase_records (
  id INTEGER PRIMARY KEY,
  roadmap_item_id INTEGER NOT NULL REFERENCES roadmap_items(id),
  phase TEXT NOT NULL CHECK(phase IN ('Intake','Explore','Define','Plan','Implement','Verify','Record')),
  revision INTEGER NOT NULL,
  content TEXT NOT NULL,
  created_at TEXT NOT NULL,
  UNIQUE(roadmap_item_id, phase, revision)
);

CREATE TABLE acceptance_criteria (
  id INTEGER PRIMARY KEY,
  roadmap_item_id INTEGER NOT NULL REFERENCES roadmap_items(id),
  title TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL CHECK(status IN ('Pending','Passed','Failed','Waived')) DEFAULT 'Pending',
  verification_method TEXT NOT NULL DEFAULT '',
  evidence TEXT NOT NULL DEFAULT '',
  verified_at TEXT
);

CREATE INDEX feature_phase_records_item_idx ON feature_phase_records(roadmap_item_id, phase, revision);
CREATE INDEX acceptance_criteria_item_idx ON acceptance_criteria(roadmap_item_id, id);
