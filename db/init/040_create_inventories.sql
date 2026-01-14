CREATE TABLE IF NOT EXISTS inventories (
  product_id uuid PRIMARY KEY REFERENCES products(id) ON DELETE CASCADE,
  quantity   integer NOT NULL DEFAULT 0,
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT chk_inventories_quantity_nonnegative CHECK (quantity >= 0)
);