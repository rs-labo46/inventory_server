import type {
  CreateProductRequest,
  CreateProductResponse,
  ListInventoriesResponse,
  ListStockMovementsResponse,
  LoginRequest,
  LoginResponse,
  StockChangeRequest,
  StockChangeResponse,
} from "./types";

const API_BASE = "";

function buildHeaders(init: RequestInit): Headers {
  const h = new Headers(init.headers);
  if (typeof init.body === "string" && !h.has("Content-Type")) {
    h.set("Content-Type", "application/json");
  }
  return h;
}

async function requestJson<TResponse>(
  path: string,
  init: RequestInit
): Promise<TResponse> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...init,
    credentials: "include", //cookie保持
    headers: buildHeaders(init),
  });

  if (!res.ok) {
    const text = await res.text();

    try {
      const parsed: unknown = JSON.parse(text);
      if (typeof parsed === "object" && parsed !== null && "error" in parsed) {
        const e = (parsed as { error?: { code?: string; message?: string } })
          .error;
        throw new Error(
          `${e?.code ?? "ERROR"}: ${e?.message ?? "unknown error"}`
        );
      }
      throw new Error(`HTTP ${res.status}: ${text}`);
    } catch {
      throw new Error(`HTTP ${res.status}: ${text}`);
    }
  }

  return (await res.json()) as TResponse;
}

export async function login(req: LoginRequest): Promise<LoginResponse> {
  return requestJson<LoginResponse>("/api/auth/login", {
    method: "POST",
    body: JSON.stringify(req),
  });
}

export async function logout(): Promise<{ ok: boolean }> {
  return requestJson<{ ok: boolean }>("/api/auth/logout", {
    method: "POST",
    body: JSON.stringify({}),
  });
}

export async function me(): Promise<LoginResponse> {
  return requestJson<LoginResponse>("/api/auth/me", { method: "GET" });
}

export async function createProduct(
  req: CreateProductRequest
): Promise<CreateProductResponse> {
  return requestJson<CreateProductResponse>("/api/products", {
    method: "POST",
    body: JSON.stringify(req),
  });
}

export async function listInventories(): Promise<ListInventoriesResponse> {
  return requestJson<ListInventoriesResponse>("/api/inventories", {
    method: "GET",
  });
}

export async function inbound(
  req: StockChangeRequest
): Promise<StockChangeResponse> {
  return requestJson<StockChangeResponse>("/api/stock/inbound", {
    method: "POST",
    body: JSON.stringify(req),
  });
}

export async function outbound(
  req: StockChangeRequest
): Promise<StockChangeResponse> {
  return requestJson<StockChangeResponse>("/api/stock/outbound", {
    method: "POST",
    body: JSON.stringify(req),
  });
}

export async function listMovements(params: {
  limit?: number;
  product_id?: string;
}): Promise<ListStockMovementsResponse> {
  const q = new URLSearchParams();

  if (typeof params.limit === "number") q.set("limit", String(params.limit));
  if (typeof params.product_id === "string" && params.product_id.length > 0) {
    q.set("product_id", params.product_id);
  }

  const suffix = q.toString().length > 0 ? `?${q.toString()}` : "";
  return requestJson<ListStockMovementsResponse>(
    `/api/stock_movements${suffix}`,
    {
      method: "GET",
    }
  );
}
