alter table public.users
  add column if not exists email_verified boolean not null default false;

update public.users
set email_verified = true
where email_verified = false;

create table if not exists public.email_verification_tokens (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null references public.users(id) on delete cascade,
  token_hash text not null unique,
  expires_at timestamptz not null,
  used_at timestamptz null,
  created_at timestamptz not null default now()
);

create index if not exists idx_email_verification_tokens_user_id
  on public.email_verification_tokens (user_id, created_at desc);

alter table public.email_verification_tokens enable row level security;

drop policy if exists "service_role_only" on public.email_verification_tokens;
create policy "service_role_only"
  on public.email_verification_tokens
  for all
  using (auth.role() = 'service_role')
  with check (auth.role() = 'service_role');
