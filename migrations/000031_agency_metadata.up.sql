ALTER TABLE users
	ADD COLUMN IF NOT EXISTS operating_country TEXT,
	ADD COLUMN IF NOT EXISTS default_currency TEXT;

-- Let's populate the existing agencies with some basic data just to make testing easier.
UPDATE users 
SET operating_country = 'Jordan', default_currency = 'JOD' 
WHERE email = 'jordan@test.com' AND operating_country IS NULL;

UPDATE users 
SET operating_country = 'Saudi Arabia', default_currency = 'SAR' 
WHERE email = 'foreign@test.com' AND operating_country IS NULL;
