package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RunMigrations menjalankan SQL schema yang diberikan user.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	stmts := []string{
		`create extension if not exists pgcrypto;`,
		`create extension if not exists citext;`,
		`do $$
begin
  if not exists (select 1 from pg_type where typname = 'user_role') then
    create type user_role as enum ('Admin', 'Employee');
  end if;
end$$;`,
		`create table if not exists public.users (
  id            uuid        primary key default gen_random_uuid(),
  name          text        not null,
  email         citext      not null unique,
  password_hash text        not null,
  role          user_role   not null default 'Employee',
  department    text,
  created_at    timestamptz not null default now()
);`,
		`create index if not exists users_created_at_idx on public.users (created_at);`,
		`create index if not exists users_department_idx on public.users (department);`,
		// tasks table
		`do $$
begin
  if not exists (select 1 from information_schema.tables where table_schema='public' and table_name='tasks') then
    create table public.tasks (
      id           uuid        primary key default gen_random_uuid(),
      user_id      uuid        not null references public.users(id) on delete cascade,
      title        text        not null,
      description  text,
      status       text        not null default 'Todo',
      due_date     timestamptz,
      created_at   timestamptz not null default now(),
      updated_at   timestamptz not null default now()
    );
  end if;
end$$;`,
		`create index if not exists tasks_user_id_idx on public.tasks (user_id);`,
		`create index if not exists tasks_status_idx on public.tasks (status);`,
		`create index if not exists tasks_due_date_idx on public.tasks (due_date);`,
	}
	sql := strings.Join(stmts, "\n")
	if _, err := pool.Exec(ctx, sql); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	return nil
}
