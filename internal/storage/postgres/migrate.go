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
	}
	sql := strings.Join(stmts, "\n")
	if _, err := pool.Exec(ctx, sql); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	return nil
}
