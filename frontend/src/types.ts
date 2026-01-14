//ログイン
export type LoginRequest = {
  email: string;
  password: string;
};

export type User = {
  id: string;
  email: string;
};

export type LoginResponse = {
  user: User;
};

//商品作成
export type CreateProductRequest = {
  name: string;
};

export type Product = {
  id: string;
  name: string;
};

export type Inventory = {
  product_id: string;
  quantity: number;
};

export type CreateProductResponse = {
  product: Product;
  inventory: Inventory;
};

//在庫一覧
export type InventoryItem = {
  product: Product;
  quantity: number;
};
export type ListInventoriesResponse = {
  items: InventoryItem[];
};

//入出庫
export type StockChangeRequest = {
  product_id: string;
  quantity: number;
  reason: string;
};
export type StockChangeResponse = {
  inventory: Inventory;
  movement: {
    id: string;
    delta: number;
    reason: string;
    created_by: string;
    created_at: string;
  };
};

//履歴
export type StockMovement = {
  id: string;
  product_id: string;
  delta: number;
  reason: string;
  created_by: string;
  created_at: string;
};

export type ListStockMovementsResponse = {
  items: StockMovement[];
};
export type ApiErrorResponse = {
  error: {
    code: string;
    message: string;
  };
};
