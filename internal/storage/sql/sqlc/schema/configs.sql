CREATE TABLE configs (
  id SERIAL PRIMARY KEY,
  website_id INTEGER NOT NULL REFERENCES websites(id) ON DELETE CASCADE,
  key TEXT NOT NULL,
  value JSON NOT NULL,
  status entity_status DEFAULT 'active' NOT NULL,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Index for filtering by status
CREATE INDEX idx_configs_status ON configs(status);