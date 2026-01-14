CREATE TABLE IF NOT EXISTS stock_movements (
  id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  product_id uuid NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  delta      integer NOT NULL,
  reason     text,
  created_by uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT chk_stock_movements_delta_nonzero CHECK (delta <> 0)
);

CREATE INDEX IF NOT EXISTS idx_stock_movements_product_created_at
  ON stock_movements(product_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_stock_movements_created_by
  ON stock_movements(created_by, created_at DESC);