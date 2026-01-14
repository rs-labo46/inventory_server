package handler

import (
	"inventory-backend/internal/application/stock"
	"inventory-backend/internal/delivery/http/responder"
	"inventory-backend/internal/domain"
	"net/http"
	"strconv"
	"time"
)

type StockMovementHandler struct {
	UC stock.MovementUsecase
}

type listMovementsRes struct {
	Items []movementItem `json:"items"`
}

type movementItem struct {
	ID        string `json:"id"`
	ProductID string `json:"product_id"`
	Delta     int    `json:"delta"`
	Reason    string `json:"reason"`
	CreatedBy string `json:"created_by"`
	CreatedAt string `json:"created_at"`
}

func (h StockMovementHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responder.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "許可されていないメソッドです")
		return
	}

	q := r.URL.Query()
	productID := q.Get("product_id")

	limit := 50
	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 || n > 200 {
			responder.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "limitは1〜200の整数で指定してください")
			return
		}
		limit = n
	}

	items, err := h.UC.List(r.Context(), productID, limit)
	if err != nil {
		if err == domain.ErrInvalid {
			responder.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "product_idが不正です")
			return
		}
		responder.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "サーバーエラーです")
		return
	}

	res := listMovementsRes{Items: make([]movementItem, 0, len(items))}
	for _, m := range items {
		res.Items = append(res.Items, movementItem{
			ID:        m.ID,
			ProductID: m.ProductID,
			Delta:     m.Delta,
			Reason:    m.Reason,
			CreatedBy: m.CreatedBy,
			CreatedAt: m.CreatedAt.Format(time.RFC3339),
		})
	}
	responder.WriteJSON(w, http.StatusOK, res)
}
