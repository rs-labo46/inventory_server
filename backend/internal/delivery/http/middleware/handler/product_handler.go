package handler

import (
	"encoding/json"
	"inventory-backend/internal/application/inventory"
	"inventory-backend/internal/delivery/http/responder"
	"inventory-backend/internal/domain"

	"net/http"
)

type ProductHandler struct {
	UC inventory.Usecase
}

type createProductReq struct {
	Name string `json:"name"`
}

type createProductRes struct {
	Product struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"product"`
	Inventory struct {
		ProductID string `json:"product_id"`
		Quantity  int    `json:"quantity"`
	} `json:"inventory"`
}

func (h ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responder.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "許可されていないメソッドです")
		return
	}

	var req createProductReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responder.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "JSONが不正です")
		return
	}
	if req.Name == "" {
		responder.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "nameは必須です")
		return
	}

	out, err := h.UC.CreateProduct(r.Context(), req.Name)
	if err != nil {
		if err == domain.ErrInvalid {
			responder.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "入力が不正です")
			return
		}
		responder.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "サーバーエラーです")
		return
	}

	var res createProductRes
	res.Product.ID = out.Product.ID
	res.Product.Name = out.Product.Name
	res.Inventory.ProductID = out.Inventory.ProductID
	res.Inventory.Quantity = out.Inventory.Quantity
	responder.WriteJSON(w, http.StatusCreated, res)
}
