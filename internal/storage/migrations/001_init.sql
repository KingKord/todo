create extension if not exists pgcrypto;

create table if not exists todos (
    id uuid primary key default gen_random_uuid(),
    title text not null,
    description text not null default '',
    completed boolean not null default false,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);
