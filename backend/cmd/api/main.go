package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"inventory-backend/internal/application/auth"
	"inventory-backend/internal/application/inventory"
	"inventory-backend/internal/application/stock"
	"inventory-backend/internal/config"
	"inventory-backend/internal/delivery/http/middleware"
	"inventory-backend/internal/delivery/http/middleware/handler"
	"inventory-backend/internal/delivery/http/responder"
	"inventory-backend/internal/infra/postgres"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type healthRes struct {
	OK bool `json:"ok"`
}

func main() {
	// config設定を読み込み
	cfg := config.MustLoad()

	// DBに接続（PostgreSQL）
	db, err := postgres.Open(cfg.DBURL)
	if err != nil {
		log.Fatalf("sql.Open failed: %v", err)
	}

	defer db.Close()

	// 起動直後にDBを確認
	if err := postgres.Ping(context.Background(), db); err != nil {
		log.Fatalf("db ping failed: %v", err)
	}

	//Repo（DBアクセスの実装）
	userRepo := postgres.UserRepo{DB: db}
	sessionRepo := postgres.SessionRepo{DB: db}
	productRepo := postgres.ProductRepo{DB: db}
	inventoryRepo := postgres.InventoryRepo{DB: db}
	stockRepo := postgres.StockRepo{DB: db}
	movementRepo := postgres.StockMovementRepo{DB: db}

	//Usecase（業務手順）
	authUC := auth.Usecase{
		Users:      userRepo,
		Sessions:   sessionRepo,
		SessionTTL: cfg.SessionTTL,
	}

	invUC := inventory.Usecase{
		Products:    productRepo,
		Inventories: inventoryRepo,
	}

	stockUC := stock.Usecase{
		Stock: stockRepo,
	}

	movementUC := stock.MovementUsecase{
		Movements: movementRepo,
	}

	//Handler
	authHandler := handler.AuthHandler{
		Auth:           authUC,
		CookieName:     cfg.SessionCookieName,
		CookieSecure:   cfg.SessionCookieSecure,
		CookieSameSite: cfg.SessionCookieSameSite,
		CookiePath:     cfg.SessionCookiePath,
	}

	productHandler := handler.ProductHandler{UC: invUC}
	inventoryHandler := handler.InventoryHandler{UC: invUC}
	stockHandler := handler.StockHandler{UC: stockUC}
	movementHandler := handler.StockMovementHandler{UC: movementUC}

	//ログイン必須
	authMW := middleware.RequireAuth{
		Sessions:   sessionRepo,
		CookieName: cfg.SessionCookieName,
	}

	// Originチェック
	originChecker := middleware.NewOriginChecker(cfg.AllowedOrigins)

	//ルーター

	root := http.NewServeMux()

	// 公開API（ログイン不要）

	// healthz公開
	root.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		responder.WriteJSON(w, http.StatusOK, healthRes{OK: true})
	})

	// login公開
	root.HandleFunc("/api/auth/login", authHandler.Login) // POST

	//認証必須API（session cookieが必要）
	protected := func(pattern string, fn http.HandlerFunc) {
		root.Handle(pattern, authMW.Wrap(http.HandlerFunc(fn)))
	}

	//logout/meはログイン後にしか使えないから保護
	protected("/api/auth/logout", authHandler.Logout) // POST
	protected("/api/auth/me", authHandler.Me)         // GET

	// 商品・在庫・在庫増減・履歴はすべてログイン必須
	protected("/api/products", productHandler.Create)       // POST
	protected("/api/inventories", inventoryHandler.List)    // GET
	protected("/api/stock/inbound", stockHandler.Inbound)   // POST
	protected("/api/stock/outbound", stockHandler.Outbound) // POST
	protected("/api/stock_movements", movementHandler.List) // GET

	//共通ミドルウェアを全体へ
	var h http.Handler = root

	// panicが起きても落とさず500を返す
	h = middleware.Recover(h)

	//リクエストIDを付与す
	h = middleware.RequestID(h)

	//Originチェック
	h = originChecker.Check(h)

	//HTTPサーバーを起動
	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("listening on %s", cfg.HTTPAddr)
	log.Fatal(srv.ListenAndServe())
}

func buildHTTPHandler(
	authHandler handler.AuthHandler,
	productHandler handler.ProductHandler,
	inventoryHandler handler.InventoryHandler,
	stockHandler handler.StockHandler,
	movementHandler handler.StockMovementHandler,
	authMW middleware.RequireAuth,
	originChecker middleware.OriginChecker,
) http.Handler {
	root := http.NewServeMux()

	// 公開API（ログイン不要）
	root.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		responder.WriteJSON(w, http.StatusOK, healthRes{OK: true})
	})
	root.HandleFunc("/api/auth/login", authHandler.Login) // POST

	// 認証必須APIsession cookie
	protected := func(pattern string, fn http.HandlerFunc) {
		root.Handle(pattern, authMW.Wrap(http.HandlerFunc(fn)))
	}

	protected("/api/auth/logout", authHandler.Logout) // POST
	protected("/api/auth/me", authHandler.Me)         // GET

	protected("/api/products", productHandler.Create)       // POST
	protected("/api/inventories", inventoryHandler.List)    // GET
	protected("/api/stock/inbound", stockHandler.Inbound)   // POST
	protected("/api/stock/outbound", stockHandler.Outbound) // POST
	protected("/api/stock_movements", movementHandler.List) // GET

	var h http.Handler = root
	h = middleware.Recover(h)
	h = middleware.RequestID(h)
	h = originChecker.Check(h)

	return h
}
