CREATE TABLE categories (
  id SERIAL PRIMARY KEY,
  website_id INTEGER NOT NULL REFERENCES websites(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  link TEXT NOT NULL UNIQUE,
  status entity_status DEFAULT 'active' NOT NULL,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Index for filtering by status
CREATE INDEX idx_categories_status ON categories(status);

-- Recommended: Index on the foreign key for faster joins
CREATE INDEX idx_categories_website_id ON categories(website_id);