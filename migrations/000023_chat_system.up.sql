create table if not exists public.chat_threads (
  id uuid primary key default gen_random_uuid(),
  pairing_id uuid not null references public.agency_pairings(id) on delete cascade,
  scope_type text not null check (scope_type in ('workspace', 'candidate')),
  candidate_id uuid null references public.candidates(id) on delete cascade,
  created_by_user_id uuid not null references public.users(id) on delete restrict,
  last_message_at timestamptz null,
  last_message_preview text null,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  constraint chk_chat_threads_scope_candidate
    check (
      (scope_type = 'workspace' and candidate_id is null)
      or
      (scope_type = 'candidate' and candidate_id is not null)
    )
);

create unique index if not exists idx_chat_threads_workspace_unique
  on public.chat_threads (pairing_id)
  where scope_type = 'workspace';

create unique index if not exists idx_chat_threads_candidate_unique
  on public.chat_threads (pairing_id, candidate_id)
  where scope_type = 'candidate';

create index if not exists idx_chat_threads_pairing_updated_desc
  on public.chat_threads (pairing_id, updated_at desc);

create index if not exists idx_chat_threads_candidate_id
  on public.chat_threads (candidate_id);

create table if not exists public.chat_messages (
  id uuid primary key default gen_random_uuid(),
  thread_id uuid not null references public.chat_threads(id) on delete cascade,
  sender_user_id uuid not null references public.users(id) on delete restrict,
  body text not null,
  created_at timestamptz not null default now()
);

create index if not exists idx_chat_messages_thread_created_desc
  on public.chat_messages (thread_id, created_at desc, id desc);

create table if not exists public.chat_thread_reads (
  id uuid primary key default gen_random_uuid(),
  thread_id uuid not null references public.chat_threads(id) on delete cascade,
  user_id uuid not null references public.users(id) on delete cascade,
  last_read_message_id uuid null references public.chat_messages(id) on delete set null,
  last_read_at timestamptz null,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  constraint uq_chat_thread_reads_thread_user unique (thread_id, user_id)
);

create index if not exists idx_chat_thread_reads_user_updated_desc
  on public.chat_thread_reads (user_id, updated_at desc);

do $$
begin
  if not exists (select 1 from pg_trigger where tgname = 'trg_chat_threads_updated_at') then
    create trigger trg_chat_threads_updated_at
    before update on public.chat_threads
    for each row
    execute function set_updated_at();
  end if;
end
$$;

do $$
begin
  if not exists (select 1 from pg_trigger where tgname = 'trg_chat_thread_reads_updated_at') then
    create trigger trg_chat_thread_reads_updated_at
    before update on public.chat_thread_reads
    for each row
    execute function set_updated_at();
  end if;
end
$$;

alter table public.chat_threads enable row level security;
alter table public.chat_messages enable row level security;
alter table public.chat_thread_reads enable row level security;

drop policy if exists "service_role_only" on public.chat_threads;
create policy "service_role_only"
  on public.chat_threads
  for all
  using (auth.role() = 'service_role')
  with check (auth.role() = 'service_role');

drop policy if exists "service_role_only" on public.chat_messages;
create policy "service_role_only"
  on public.chat_messages
  for all
  using (auth.role() = 'service_role')
  with check (auth.role() = 'service_role');

drop policy if exists "service_role_only" on public.chat_thread_reads;
create policy "service_role_only"
  on public.chat_thread_reads
  for all
  using (auth.role() = 'service_role')
  with check (auth.role() = 'service_role');
