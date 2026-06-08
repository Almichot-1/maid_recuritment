-- Seed test users for local development
-- Password for both users: "password123"
-- This is a bcrypt hash of "password123"

-- Ethiopian Agent (creates candidates)
INSERT INTO users (
    id,
    email,
    password_hash,
    full_name,
    role,
    company_name,
    is_active,
    created_at,
    updated_at
) VALUES (
    '11111111-1111-1111-1111-111111111111',
    'ethiopian@test.com',
    '$2a$10$rKJ5VlZVlZVlZVlZVlZVluXJ5YvYvYvYvYvYvYvYvYvYvYvYvYvYu',
    'Ethiopian Test Agent',
    'ethiopian_agent',
    'Ethiopian Recruitment Agency',
    true,
    NOW(),
    NOW()
) ON CONFLICT (email) DO NOTHING;

-- Foreign Agent (selects candidates)
INSERT INTO users (
    id,
    email,
    password_hash,
    full_name,
    role,
    company_name,
    is_active,
    created_at,
    updated_at
) VALUES (
    '22222222-2222-2222-2222-222222222222',
    'foreign@test.com',
    '$2a$10$rKJ5VlZVlZVlZVlZVlZVluXJ5YvYvYvYvYvYvYvYvYvYvYvYvYvYu',
    'Foreign Test Agent',
    'foreign_agent',
    'International Recruitment Co',
    true,
    NOW(),
    NOW()
);

-- Display created users
SELECT 
    email,
    role,
    full_name,
    company_name,
    is_active
FROM users
ORDER BY role;
