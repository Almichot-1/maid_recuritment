drop policy if exists "service_role_only" on public.email_verification_tokens;
drop table if exists public.email_verification_tokens;
alter table public.users drop column if exists email_verified;

