CREATE TABLE products (
  id SERIAL PRIMARY KEY,
  website_id INTEGER NOT NULL REFERENCES websites(id) ON DELETE CASCADE,
  category_id INTEGER NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
  title TEXT NOT NULL,
  price BIGINT NOT NULL,
  link TEXT NOT NULL UNIQUE,
  image TEXT,
  description TEXT,
  status entity_status DEFAULT 'active' NOT NULL,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Index foreign keys for faster JOINS and CASCADE deletions
CREATE INDEX idx_products_website_id ON products(website_id);
CREATE INDEX idx_products_category_id ON products(category_id);

-- Index for filtering by status
CREATE INDEX idx_products_status ON products(status);