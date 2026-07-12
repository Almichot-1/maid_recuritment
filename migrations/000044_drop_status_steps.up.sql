-- Migration: Drop old status_steps table (replaced by selection_progress)

DROP TABLE IF EXISTS status_steps CASCADE;
