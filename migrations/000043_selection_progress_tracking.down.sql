-- Migration rollback: Drop selection progress tracking table

DROP TABLE IF EXISTS selection_progress CASCADE;
