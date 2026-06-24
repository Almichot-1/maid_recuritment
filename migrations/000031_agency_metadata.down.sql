ALTER TABLE users
	DROP COLUMN IF EXISTS operating_country,
	DROP COLUMN IF EXISTS default_currency;
