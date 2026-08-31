CREATE TABLE feature_reports (
  id INTEGER PRIMARY KEY,
  roadmap_item_id INTEGER NOT NULL REFERENCES roadmap_items(id),
  execution_mode TEXT NOT NULL,
  roles TEXT NOT NULL,
  elapsed_ns INTEGER NOT NULL,
  tool_calls INTEGER NOT NULL,
  verification TEXT NOT NULL,
  input_tokens INTEGER,
  cached_input_tokens INTEGER,
  output_tokens INTEGER,
  reasoning_tokens INTEGER,
  total_tokens INTEGER,
  created_at TEXT NOT NULL
);
