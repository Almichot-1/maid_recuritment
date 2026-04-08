drop policy if exists "service_role_only" on public.chat_thread_reads;
drop policy if exists "service_role_only" on public.chat_messages;
drop policy if exists "service_role_only" on public.chat_threads;

drop table if exists public.chat_thread_reads;
drop table if exists public.chat_messages;
drop table if exists public.chat_threads;
