ALTER TABLE roadmap_items ADD COLUMN feature_type TEXT NOT NULL DEFAULT 'Change' CHECK(feature_type IN ('Capability', 'Change', 'Gap'));
ALTER TABLE roadmap_items ADD COLUMN current_state TEXT NOT NULL DEFAULT '';

CREATE TABLE feature_relationships (
    id INTEGER PRIMARY KEY,
    source_item_id INTEGER NOT NULL REFERENCES roadmap_items(id),
    target_item_id INTEGER NOT NULL REFERENCES roadmap_items(id),
    relationship_type TEXT NOT NULL CHECK(relationship_type IN ('extends', 'depends_on', 'replaces')),
    created_at TEXT NOT NULL,
    CHECK(source_item_id != target_item_id),
    UNIQUE(source_item_id, target_item_id, relationship_type)
);

CREATE INDEX feature_relationships_source_idx ON feature_relationships(source_item_id);
CREATE INDEX feature_relationships_target_idx ON feature_relationships(target_item_id);
