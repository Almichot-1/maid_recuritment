DO $$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'agency_pairing_status') THEN
		CREATE TYPE agency_pairing_status AS ENUM ('active', 'suspended', 'ended');
	END IF;
END
$$;

CREATE TABLE IF NOT EXISTS agency_pairings (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	ethiopian_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	foreign_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	status agency_pairing_status NOT NULL DEFAULT 'active',
	approved_by_admin_id UUID REFERENCES admins(id) ON DELETE SET NULL,
	approved_at TIMESTAMPTZ,
	ended_at TIMESTAMPTZ,
	notes TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS candidate_pair_shares (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	pairing_id UUID NOT NULL REFERENCES agency_pairings(id) ON DELETE CASCADE,
	candidate_id UUID NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
	shared_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
	is_active BOOLEAN NOT NULL DEFAULT TRUE,
	shared_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	unshared_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE selections
	ADD COLUMN IF NOT EXISTS pairing_id UUID;

CREATE INDEX IF NOT EXISTS idx_agency_pairings_ethiopian_user_id ON agency_pairings(ethiopian_user_id);
CREATE INDEX IF NOT EXISTS idx_agency_pairings_foreign_user_id ON agency_pairings(foreign_user_id);
CREATE INDEX IF NOT EXISTS idx_agency_pairings_status ON agency_pairings(status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_agency_pairings_unique_active_pair
	ON agency_pairings(ethiopian_user_id, foreign_user_id)
	WHERE status = 'active';

CREATE INDEX IF NOT EXISTS idx_candidate_pair_shares_pairing_id ON candidate_pair_shares(pairing_id);
CREATE INDEX IF NOT EXISTS idx_candidate_pair_shares_candidate_id ON candidate_pair_shares(candidate_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_candidate_pair_shares_unique_active
	ON candidate_pair_shares(pairing_id, candidate_id)
	WHERE is_active = TRUE;

DO $$
BEGIN
	IF NOT EXISTS (
		SELECT 1
		FROM information_schema.table_constraints
		WHERE constraint_name = 'fk_selections_pairing_id'
			AND table_name = 'selections'
	) THEN
		ALTER TABLE selections
			ADD CONSTRAINT fk_selections_pairing_id
			FOREIGN KEY (pairing_id) REFERENCES agency_pairings(id) ON DELETE RESTRICT;
	END IF;
END
$$;

INSERT INTO agency_pairings (
	id,
	ethiopian_user_id,
	foreign_user_id,
	status,
	approved_at,
	created_at,
	updated_at
)
SELECT
	gen_random_uuid(),
	source.ethiopian_user_id,
	source.foreign_user_id,
	'active'::agency_pairing_status,
	COALESCE(source.first_selected_at, NOW()),
	COALESCE(source.first_selected_at, NOW()),
	COALESCE(source.last_updated_at, NOW())
FROM (
	SELECT
		c.created_by AS ethiopian_user_id,
		s.selected_by AS foreign_user_id,
		MIN(s.created_at) AS first_selected_at,
		MAX(COALESCE(s.updated_at, s.created_at)) AS last_updated_at
	FROM selections s
	JOIN candidates c ON c.id = s.candidate_id
	GROUP BY c.created_by, s.selected_by
) source
LEFT JOIN agency_pairings ap
	ON ap.ethiopian_user_id = source.ethiopian_user_id
	AND ap.foreign_user_id = source.foreign_user_id
	AND ap.status = 'active'
WHERE ap.id IS NULL;

UPDATE selections s
SET pairing_id = ap.id
FROM candidates c,
	agency_pairings ap
WHERE c.id = s.candidate_id
	AND ap.ethiopian_user_id = c.created_by
	AND ap.foreign_user_id = s.selected_by
	AND ap.status = 'active'
	AND s.pairing_id IS NULL;

DO $$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_agency_pairings_updated_at') THEN
		CREATE TRIGGER trg_agency_pairings_updated_at
		BEFORE UPDATE ON agency_pairings
		FOR EACH ROW
		EXECUTE FUNCTION set_updated_at();
	END IF;
END
$$;

DO $$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_candidate_pair_shares_updated_at') THEN
		CREATE TRIGGER trg_candidate_pair_shares_updated_at
		BEFORE UPDATE ON candidate_pair_shares
		FOR EACH ROW
		EXECUTE FUNCTION set_updated_at();
	END IF;
END
$$;
