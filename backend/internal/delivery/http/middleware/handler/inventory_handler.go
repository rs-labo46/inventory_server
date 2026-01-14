package handler

import (
	"inventory-backend/internal/application/inventory"
	"inventory-backend/internal/delivery/http/responder"
	"net/http"
)

type InventoryHandler struct {
	UC inventory.Usecase
}

type listInventoriesRes struct {
	Items []inventoryItem `json:"items"`
}

type inventoryItem struct {
	Product struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"product"`
	Quantity int `json:"quantity"`
}

func (h InventoryHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responder.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "許可されていないメソッドです")
		return
	}

	items, err := h.UC.ListInventories(r.Context())
	if err != nil {
		responder.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "サーバーエラーです")
		return
	}
	res := listInventoriesRes{
		Items: make([]inventoryItem, 0, len(items)),
	}

	for _, it := range items {
		var v inventoryItem
		v.Product.ID = it.ProductID
		v.Product.Name = it.ProductName
		v.Quantity = it.Quantity
		res.Items = append(res.Items, v)
	}
	responder.WriteJSON(w, http.StatusOK, res)
}
