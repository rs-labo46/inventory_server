import { useEffect, useMemo, useState } from "react";
import {
  inbound,
  listInventories,
  listMovements,
  login,
  logout,
  me,
  outbound,
  createProduct,
} from "../api";
import type { InventoryItem, StockMovement, User } from "../types";

type Status = {
  message: string;
  kind: "idle" | "success" | "error";
};

export default function App() {
  const [user, setUser] = useState<User | null>(null);

  const [status, setStatus] = useState<Status>({ message: "", kind: "idle" });

  const [email, setEmail] = useState<string>("admin@example.com");
  const [password, setPassword] = useState<string>("admin0001");

  const [productName, setProductName] = useState<string>("");

  const [inventories, setInventories] = useState<InventoryItem[]>([]);

  const [selectedProductId, setSelectedProductId] = useState<string>("");
  const [qty, setQty] = useState<number>(1);
  const [reason, setReason] = useState<string>("");

  const [movements, setMovements] = useState<StockMovement[]>([]);
  const [movementLimit, setMovementLimit] = useState<number>(50);
  const [movementFilterProductId, setMovementFilterProductId] =
    useState<string>("");

  const productOptions = useMemo(
    () => inventories.map((i) => i.product),
    [inventories]
  );

  useEffect(() => {
    void (async () => {
      try {
        const res = await me();
        setUser(res.user);
        setStatus({ message: "ログイン状態を復元しました", kind: "success" });
        await refreshAll();
      } catch {}
    })();
  }, []);

  async function refreshAll(): Promise<void> {
    await Promise.all([refreshInventories(), refreshMovements()]);
  }

  async function refreshInventories(): Promise<void> {
    const res = await listInventories();
    setInventories(res.items);

    if (res.items.length > 0 && selectedProductId === "") {
      setSelectedProductId(res.items[0]!.product.id);
    }
  }

  async function refreshMovements(): Promise<void> {
    const res = await listMovements({
      limit: movementLimit,
      product_id:
        movementFilterProductId.length > 0
          ? movementFilterProductId
          : undefined,
    });
    setMovements(res.items);
  }

  async function onLogin(): Promise<void> {
    setStatus({ message: "", kind: "idle" });

    try {
      const res = await login({ email, password });
      setUser(res.user);
      setStatus({ message: "ログインしました", kind: "success" });
      await refreshAll();
    } catch (e) {
      setStatus({ message: String(e), kind: "error" });
    }
  }

  async function onLogout(): Promise<void> {
    setStatus({ message: "", kind: "idle" });

    try {
      await logout();
      setUser(null);
      setInventories([]);
      setMovements([]);
      setStatus({ message: "ログアウトしました", kind: "success" });
    } catch (e) {
      setStatus({ message: String(e), kind: "error" });
    }
  }

  async function onCreateProduct(): Promise<void> {
    setStatus({ message: "", kind: "idle" });

    if (productName.trim().length === 0) {
      setStatus({ message: "商品名を入力してください", kind: "error" });
      return;
    }

    try {
      await createProduct({ name: productName.trim() });
      setProductName("");
      setStatus({ message: "商品を作成しました", kind: "success" });
      await refreshInventories();
    } catch (e) {
      setStatus({ message: String(e), kind: "error" });
    }
  }

  async function onInbound(): Promise<void> {
    setStatus({ message: "", kind: "idle" });

    if (selectedProductId.length === 0) {
      setStatus({ message: "product_id を選んでください", kind: "error" });
      return;
    }
    if (qty <= 0) {
      setStatus({ message: "quantity は 1 以上にしてください", kind: "error" });
      return;
    }

    try {
      await inbound({
        product_id: selectedProductId,
        quantity: qty,
        reason,
      });
      setStatus({ message: "入庫しました", kind: "success" });
      await refreshAll();
    } catch (e) {
      setStatus({ message: String(e), kind: "error" });
    }
  }

  async function onOutbound(): Promise<void> {
    setStatus({ message: "", kind: "idle" });

    if (selectedProductId.length === 0) {
      setStatus({ message: "product_id を選んでください", kind: "error" });
      return;
    }
    if (qty <= 0) {
      setStatus({ message: "quantity は 1 以上にしてください", kind: "error" });
      return;
    }

    try {
      await outbound({
        product_id: selectedProductId,
        quantity: qty,
        reason,
      });
      setStatus({ message: "出庫しました", kind: "success" });
      await refreshAll();
    } catch (e) {
      setStatus({ message: String(e), kind: "error" });
    }
  }

  return (
    <div
      style={{
        maxWidth: 980,
        margin: "0 auto",
        padding: 16,
        fontFamily: "system-ui",
      }}
    >
      <h1>Inventory Server</h1>

      {status.message.length > 0 && (
        <div
          style={{
            padding: 12,
            border: "1px solid #ddd",
            borderRadius: 8,
            marginBottom: 16,
          }}
        >
          <b>{status.kind.toUpperCase()}</b>: {status.message}
        </div>
      )}

      <section
        style={{
          padding: 12,
          border: "1px solid #ddd",
          borderRadius: 8,
          marginBottom: 16,
        }}
      >
        <h2>Login</h2>

        {user ? (
          <div>
            <div>Logged in as: {user.email}</div>
            <button onClick={() => void onLogout()}>Logout</button>
          </div>
        ) : (
          <div style={{ display: "grid", gap: 8 }}>
            <label>
              Email
              <input
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                style={{ width: "100%" }}
              />
            </label>
            <label>
              Password
              <input
                value={password}
                type="password"
                onChange={(e) => setPassword(e.target.value)}
                style={{ width: "100%" }}
              />
            </label>
            <button onClick={() => void onLogin()}>Login</button>
          </div>
        )}
      </section>

      <section
        style={{
          padding: 12,
          border: "1px solid #ddd",
          borderRadius: 8,
          marginBottom: 16,
        }}
      >
        <h2>商品追加</h2>
        <div style={{ display: "flex", gap: 8 }}>
          <input
            placeholder="name"
            value={productName}
            onChange={(e) => setProductName(e.target.value)}
            style={{ flex: 1 }}
          />
          <button disabled={!user} onClick={() => void onCreateProduct()}>
            Create
          </button>
        </div>
        {!user && (
          <div style={{ marginTop: 8 }}>※ ログイン後に操作できます</div>
        )}
      </section>

      <section
        style={{
          padding: 12,
          border: "1px solid #ddd",
          borderRadius: 8,
          marginBottom: 16,
        }}
      >
        <h2>在庫一覧</h2>
        <button disabled={!user} onClick={() => void refreshInventories()}>
          在庫一覧を更新
        </button>
        <div style={{ marginTop: 12, overflowX: "auto" }}>
          <table style={{ width: "100%", borderCollapse: "collapse" }}>
            <thead>
              <tr>
                <th
                  style={{ textAlign: "left", borderBottom: "1px solid #ddd" }}
                >
                  product_id
                </th>
                <th
                  style={{ textAlign: "left", borderBottom: "1px solid #ddd" }}
                >
                  name
                </th>
                <th
                  style={{ textAlign: "right", borderBottom: "1px solid #ddd" }}
                >
                  quantity
                </th>
              </tr>
            </thead>
            <tbody>
              {inventories.map((it) => (
                <tr key={it.product.id}>
                  <td style={{ borderBottom: "1px solid #f0f0f0" }}>
                    {it.product.id}
                  </td>
                  <td style={{ borderBottom: "1px solid #f0f0f0" }}>
                    {it.product.name}
                  </td>
                  <td
                    style={{
                      borderBottom: "1px solid #f0f0f0",
                      textAlign: "right",
                    }}
                  >
                    {it.quantity}
                  </td>
                </tr>
              ))}
              {inventories.length === 0 && (
                <tr>
                  <td colSpan={3} style={{ padding: 8 }}>
                    アイテムがありません。
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </section>

      <section
        style={{
          padding: 12,
          border: "1px solid #ddd",
          borderRadius: 8,
          marginBottom: 16,
        }}
      >
        <h2>入庫 / 出庫</h2>

        <div style={{ display: "grid", gap: 8, maxWidth: 520 }}>
          <label>
            Product
            <select
              value={selectedProductId}
              onChange={(e) => setSelectedProductId(e.target.value)}
              style={{ width: "100%" }}
              disabled={!user}
            >
              {productOptions.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name} ({p.id})
                </option>
              ))}
              {productOptions.length === 0 && (
                <option value="">(no products)</option>
              )}
            </select>
          </label>

          <label>
            Quantity
            <input
              type="number"
              value={qty}
              min={1}
              onChange={(e) => setQty(Number(e.target.value))}
              disabled={!user}
              style={{ width: "100%" }}
            />
          </label>

          <label>
            Reason
            <input
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              disabled={!user}
              style={{ width: "100%" }}
            />
          </label>

          <div style={{ display: "flex", gap: 8 }}>
            <button disabled={!user} onClick={() => void onInbound()}>
              Inbound (+)
            </button>
            <button disabled={!user} onClick={() => void onOutbound()}>
              Outbound (-)
            </button>
          </div>
        </div>
      </section>

      <section
        style={{ padding: 12, border: "1px solid #ddd", borderRadius: 8 }}
      >
        <h2>履歴</h2>

        <div
          style={{
            display: "flex",
            gap: 8,
            flexWrap: "wrap",
            alignItems: "center",
          }}
        >
          <label>
            limit
            <input
              type="number"
              min={1}
              max={200}
              value={movementLimit}
              onChange={(e) => setMovementLimit(Number(e.target.value))}
              disabled={!user}
              style={{ width: 120, marginLeft: 6 }}
            />
          </label>

          <label>
            product_idフィルター
            <input
              value={movementFilterProductId}
              onChange={(e) => setMovementFilterProductId(e.target.value)}
              disabled={!user}
              style={{ width: 420, marginLeft: 6 }}
            />
          </label>

          <button disabled={!user} onClick={() => void refreshMovements()}>
            履歴を更新する
          </button>
        </div>

        <div style={{ marginTop: 12, overflowX: "auto" }}>
          <table style={{ width: "100%", borderCollapse: "collapse" }}>
            <thead>
              <tr>
                <th
                  style={{ textAlign: "left", borderBottom: "1px solid #ddd" }}
                >
                  created_at
                </th>
                <th
                  style={{ textAlign: "left", borderBottom: "1px solid #ddd" }}
                >
                  product_id
                </th>
                <th
                  style={{ textAlign: "right", borderBottom: "1px solid #ddd" }}
                >
                  delta
                </th>
                <th
                  style={{ textAlign: "left", borderBottom: "1px solid #ddd" }}
                >
                  reason
                </th>
              </tr>
            </thead>
            <tbody>
              {movements.map((m) => (
                <tr key={m.id}>
                  <td style={{ borderBottom: "1px solid #f0f0f0" }}>
                    {m.created_at}
                  </td>
                  <td style={{ borderBottom: "1px solid #f0f0f0" }}>
                    {m.product_id}
                  </td>
                  <td
                    style={{
                      borderBottom: "1px solid #f0f0f0",
                      textAlign: "right",
                    }}
                  >
                    {m.delta}
                  </td>
                  <td style={{ borderBottom: "1px solid #f0f0f0" }}>
                    {m.reason}
                  </td>
                </tr>
              ))}
              {movements.length === 0 && (
                <tr>
                  <td colSpan={4} style={{ padding: 8 }}>
                    アイテムがありません。
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </section>
    </div>
  );
}
