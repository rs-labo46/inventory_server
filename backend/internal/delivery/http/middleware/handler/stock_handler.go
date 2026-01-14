package handler

import (
	"encoding/json"
	"errors"
	"inventory-backend/internal/application/stock"
	"inventory-backend/internal/delivery/http/middleware"
	"inventory-backend/internal/delivery/http/responder"
	"inventory-backend/internal/domain"

	"net/http"
	"time"
)

type StockHandler struct {
	UC stock.Usecase
}

type changeReq struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Reason    string `json:"reason"`
}

type changeRes struct {
	Inventory struct {
		ProductID string `json:"product_id"`
		Quantity  int    `json:"quantity"`
	} `json:"inventory"`

	Movement struct {
		ID        string `json:"id"`
		Delta     int    `json:"delta"`
		Reason    string `json:"reason"`
		CreatedBy string `json:"created_by"`
		CreatedAt string `json:"created_at"`
	} `json:"movement"`
}

func (h StockHandler) Inbound(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responder.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "許可されていないメソッドです")
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		responder.WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "ログインしてください")
		return
	}

	var req changeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responder.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "JSONが不正です")
		return
	}

	out, err := h.UC.Inbound(r.Context(), userID, req.ProductID, req.Quantity, req.Reason)
	if err != nil {
		writeStockErr(w, err)
		return
	}

	var res changeRes
	res.Inventory.ProductID = out.Inventory.ProductID
	res.Inventory.Quantity = out.Inventory.Quantity
	res.Movement.ID = out.Movement.ID
	res.Movement.Delta = out.Movement.Delta
	res.Movement.Reason = out.Movement.Reason
	res.Movement.CreatedBy = out.Movement.CreatedBy
	res.Movement.CreatedAt = out.Movement.CreatedAt.Format(time.RFC3339)

	responder.WriteJSON(w, http.StatusOK, res)
}

func (h StockHandler) Outbound(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responder.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "許可されていないメソッドです")
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		responder.WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "ログインしてください")
		return
	}

	var req changeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responder.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "JSONが不正です")
		return
	}

	out, err := h.UC.Outbound(r.Context(), userID, req.ProductID, req.Quantity, req.Reason)
	if err != nil {
		writeStockErr(w, err)
		return
	}

	var res changeRes
	res.Inventory.ProductID = out.Inventory.ProductID
	res.Inventory.Quantity = out.Inventory.Quantity
	res.Movement.ID = out.Movement.ID
	res.Movement.Delta = out.Movement.Delta
	res.Movement.Reason = out.Movement.Reason
	res.Movement.CreatedBy = out.Movement.CreatedBy
	res.Movement.CreatedAt = out.Movement.CreatedAt.Format(time.RFC3339)

	responder.WriteJSON(w, http.StatusOK, res)
}

func writeStockErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalid):
		responder.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "入力が不正です")
	case errors.Is(err, domain.ErrNotFound):
		responder.WriteError(w, http.StatusNotFound, "NOT_FOUND", "対象が見つかりません")
	case errors.Is(err, domain.ErrInsufficientStock):
		responder.WriteError(w, http.StatusConflict, "INSUFFICIENT_STOCK", "在庫が足りません")
	default:
		responder.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "サーバーエラーです")
	}
}
