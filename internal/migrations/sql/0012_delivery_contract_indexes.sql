CREATE UNIQUE INDEX acceptance_criteria_plan_key_idx ON acceptance_criteria(plan_revision_id, criterion_key) WHERE plan_revision_id IS NOT NULL;
