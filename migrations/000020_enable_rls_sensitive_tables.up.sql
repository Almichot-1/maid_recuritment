-- Enable RLS for sensitive/public tables to prevent anonymous API access
alter table public.user_sessions enable row level security;
alter table public.admin_setup_tokens enable row level security;
alter table public.password_reset_requests enable row level security;
alter table public.medical_data enable row level security;
alter table public.passport_data enable row level security;

