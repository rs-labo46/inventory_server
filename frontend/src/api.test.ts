import { describe, it, expect, beforeEach, vi } from "vitest";
import {
  login,
  logout,
  me,
  createProduct,
  listInventories,
  inbound,
  outbound,
  listMovements,
} from "./api";
import type {
  LoginResponse,
  CreateProductResponse,
  StockChangeResponse,
  ListStockMovementsResponse,
} from "./types";

function jsonResponse<T>(body: T, init?: ResponseInit): Response {
  return new Response(JSON.stringify(body), {
    status: init?.status ?? 200,
    headers: {
      "Content-Type": "application/json; charset=utf-8",
      ...(init?.headers ?? {}),
    },
  });
}

function textResponse(text: string, init?: ResponseInit): Response {
  return new Response(text, {
    status: init?.status ?? 500,
    headers: {
      "Content-Type": "text/plain; charset=utf-8",
      ...(init?.headers ?? {}),
    },
  });
}

describe("src/api.ts", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it("login: credentials include + Content-Type 自動付与 + JSON成功", async () => {
    const mockLoginRes: LoginResponse = {
      user: { id: "u1", email: "admin@example.com" },
    };

    const spy = vi
      .spyOn(globalThis, "fetch")
      .mockResolvedValue(jsonResponse(mockLoginRes, { status: 200 }));

    const got = await login({
      email: "admin@example.com",
      password: "admin0001",
    });
    expect(got).toEqual(mockLoginRes);

    expect(spy).toHaveBeenCalledTimes(1);

    const [url, init] = spy.mock.calls[0];
    expect(String(url)).toBe("/api/auth/login");
    expect(init?.credentials).toBe("include");

    const h = new Headers(init?.headers);
    expect(h.get("Content-Type")).toBe("application/json");
  });

  it("me: GET なので Content-Type は付かない（bodyが無いから）", async () => {
    const mockMeRes: LoginResponse = {
      user: { id: "u1", email: "admin@example.com" },
    };

    const spy = vi
      .spyOn(globalThis, "fetch")
      .mockResolvedValue(jsonResponse(mockMeRes, { status: 200 }));

    const got = await me();
    expect(got).toEqual(mockMeRes);

    const [, init] = spy.mock.calls[0];
    const h = new Headers(init?.headers);
    expect(h.has("Content-Type")).toBe(false);
  });

  it("res.ok=false で error JSON を返すと 'CODE: message' になる", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      jsonResponse(
        { error: { code: "UNAUTHENTICATED", message: "ログインしてください" } },
        { status: 401 }
      )
    );

    await expect(me()).rejects.toThrowError(
      "UNAUTHENTICATED: ログインしてください"
    );
  });

  it("res.ok=false で JSONじゃない本文なら 'HTTP <status>: <text>' になる", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      textResponse("oops", { status: 500 })
    );

    await expect(listInventories()).rejects.toThrowError("HTTP 500: oops");
  });

  it("createProduct / logout / inbound / outbound は body があるので Content-Type が付く", async () => {
    const spy = vi.spyOn(globalThis, "fetch");

    const mockCreateRes: CreateProductResponse = {
      product: { id: "p1", name: "coffee" },
      inventory: { product_id: "p1", quantity: 0 },
    };
    spy.mockResolvedValueOnce(jsonResponse(mockCreateRes, { status: 201 }));
    await createProduct({ name: "coffee" });

    spy.mockResolvedValueOnce(jsonResponse({ ok: true }, { status: 200 }));
    await logout();

    const mockStockRes: StockChangeResponse = {
      inventory: { product_id: "p1", quantity: 10 },
      movement: {
        id: "m1",
        delta: 10,
        reason: "in",
        created_by: "u1",
        created_at: new Date().toISOString(),
      },
    };
    spy.mockResolvedValueOnce(jsonResponse(mockStockRes, { status: 200 }));
    await inbound({ product_id: "p1", quantity: 10, reason: "in" });

    spy.mockResolvedValueOnce(jsonResponse(mockStockRes, { status: 200 }));
    await outbound({ product_id: "p1", quantity: 1, reason: "out" });

    expect(spy).toHaveBeenCalledTimes(4);

    for (const [, init] of spy.mock.calls) {
      const h = new Headers(init?.headers);
      expect(h.get("Content-Type")).toBe("application/json");
      expect(init?.credentials).toBe("include");
    }
  });

  it("listMovements: クエリ無しなら /api/stock_movements", async () => {
    const mock: ListStockMovementsResponse = { items: [] };

    const spy = vi
      .spyOn(globalThis, "fetch")
      .mockResolvedValue(jsonResponse(mock, { status: 200 }));

    await listMovements({});

    const [url] = spy.mock.calls[0];
    expect(String(url)).toBe("/api/stock_movements");
  });

  it("listMovements: limit と product_id の両方が付く", async () => {
    const mock: ListStockMovementsResponse = { items: [] };

    const spy = vi
      .spyOn(globalThis, "fetch")
      .mockResolvedValue(jsonResponse(mock, { status: 200 }));

    await listMovements({ limit: 10, product_id: "p1" });

    const [url] = spy.mock.calls[0];
    const s = String(url);
    expect(s.startsWith("/api/stock_movements?")).toBe(true);
    expect(s.includes("limit=10")).toBe(true);
    expect(s.includes("product_id=p1")).toBe(true);
  });
});
