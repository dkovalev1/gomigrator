--gomigrator
create schema if not exists gomigrator;

--gomigrator
create table gomigrator.migrations (
    id serial primary key,
    name text,
    status int,
    last_run timestamp
);
