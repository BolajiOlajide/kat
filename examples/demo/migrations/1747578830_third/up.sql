CREATE EXTENSION IF NOT EXISTS "uuid-ossp";  -- PostgreSQL-specific for UUID generation

CREATE TABLE projects (
    project_id   UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_name TEXT        NOT NULL,
    start_date   DATE,
    end_date     DATE,
    budget       DECIMAL(15,2) CHECK (budget >= 0)
);
