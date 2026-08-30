ALTER TABLE roadmap_items ADD COLUMN target_date TEXT NOT NULL DEFAULT '';
ALTER TABLE roadmap_items ADD COLUMN progress INTEGER NOT NULL DEFAULT 0 CHECK(progress >= 0 AND progress <= 100);
ALTER TABLE roadmap_items ADD COLUMN priority TEXT NOT NULL DEFAULT 'Medium' CHECK(priority IN ('High', 'Medium', 'Low'));
ALTER TABLE roadmap_items ADD COLUMN complexity TEXT NOT NULL DEFAULT 'Medium' CHECK(complexity IN ('High', 'Medium', 'Low'));
ALTER TABLE roadmap_items ADD COLUMN team TEXT NOT NULL DEFAULT '';
ALTER TABLE roadmap_items ADD COLUMN lead_engineer TEXT NOT NULL DEFAULT '';
ALTER TABLE roadmap_items ADD COLUMN technical_summary TEXT NOT NULL DEFAULT '';
ALTER TABLE roadmap_items ADD COLUMN specifications TEXT NOT NULL DEFAULT '[]';
ALTER TABLE roadmap_items ADD COLUMN documentation_links TEXT NOT NULL DEFAULT '[]';
