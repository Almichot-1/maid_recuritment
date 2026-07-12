-- 000039_candidate_experience_entries.up.sql
-- ExperienceAbroad column already exists as TEXT, which can hold JSON.
-- The Go code now reads/writes it as []ExperienceEntry JSON array.
-- No DDL change needed.
SELECT 1;
