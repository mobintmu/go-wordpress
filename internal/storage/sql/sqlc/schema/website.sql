CREATE TYPE entity_status AS ENUM ('active', 'inactive', 'in_progress');

CREATE TABLE websites (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    domain TEXT NOT NULL UNIQUE,
    status entity_status DEFAULT 'active' NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Create index on status for filtering
CREATE INDEX idx_websites_status ON websites(status);