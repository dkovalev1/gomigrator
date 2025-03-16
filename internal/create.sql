--gomigrator
CREATE SCHEMA IF NOT EXISTS gomigrator;

--gomigrator
CREATE TYPE migration_status AS ENUM ('new', 'inprogress', 'error', 'applied');
CREATE TYPE migration_type AS ENUM ('go', 'sql');

--gomigrator
CREATE TABLE gomigrator.migrations (
    id SERIAL,
    name text primary key,
    type migration_type,
    status migration_status DEFAULT 'created',
    last_run timestamp
);
