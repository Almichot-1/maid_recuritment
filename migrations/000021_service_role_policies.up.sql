-- Explicitly restrict all public tables to service role only.
-- This removes "RLS enabled but no policy" warnings and blocks anon/auth access.

do $$
declare
  tbl text;
  tables text[] := array[
    'admins',
    'admin_setup_tokens',
    'agency_approval_requests',
    'agency_pairings',
    'approvals',
    'audit_logs',
    'candidate_pair_shares',
    'candidates',
    'documents',
    'medical_data',
    'notifications',
    'passport_data',
    'password_reset_requests',
    'platform_settings',
    'selections',
    'status_steps',
    'user_sessions',
    'users'
  ];
begin
  foreach tbl in array tables loop
    execute format('drop policy if exists "service_role_only" on public.%I', tbl);
    execute format(
      'create policy "service_role_only" on public.%I for all using (auth.role() = ''service_role'') with check (auth.role() = ''service_role'')',
      tbl
    );
  end loop;
end $$;

