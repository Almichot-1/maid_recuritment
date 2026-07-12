UPDATE candidates SET experience_abroad = '[]'::jsonb WHERE experience_abroad IS NULL OR experience_abroad = '';
ALTER TABLE candidates ALTER COLUMN experience_abroad TYPE JSONB USING experience_abroad::jsonb;
ALTER TABLE candidates ALTER COLUMN experience_abroad SET DEFAULT '[]'::jsonb;
ALTER TABLE candidates ALTER COLUMN experience_abroad SET NOT NULL;
