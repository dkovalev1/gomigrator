--gomigrator up
CREATE SCHEMA IF NOT EXISTS gomigrator;

--gomigrator up
CREATE TYPE migration_status AS ENUM ('new', 'inprogress', 'error', 'applied');
CREATE TYPE migration_type AS ENUM ('go', 'sql');

--gomigrator up
CREATE TABLE gomigrator.migrations (
    mid SERIAL,
    mname text primary key,
    mtype migration_type,
    mstatus migration_status DEFAULT 'new',
    mlastrun timestamp
);
