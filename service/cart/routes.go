package cart

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/thomzes/ecom-go/service/auth"
	"github.com/thomzes/ecom-go/types"
	"github.com/thomzes/ecom-go/utils"
)

type Handler struct {
	userStore  types.UserStore
	store      types.ProductStore
	orderStore types.OrderStore
}

func NewHandler(store types.ProductStore, userStore types.UserStore, orderStore types.OrderStore) *Handler {
	return &Handler{store: store, userStore: userStore, orderStore: orderStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/cart/checkout", auth.WithJWTAuth(h.handleCheckout, h.userStore)).Methods(http.MethodPost)
	// router.HandleFunc("/cart/checkout/{id}", auth.WithJWTAuth(h.handleUpdateCheckout)).Methods(http.MethodPut)
}

func (h *Handler) handleCheckout(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())

	var cart types.CartCheckoutPayload
	if err := utils.ParseJSON(r, &cart); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := utils.Validate.Struct(cart); err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", err))
		return
	}

	productIds, err := getCartItemsIDs(cart.Items)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// get products
	prodcuts, err := h.store.GetProductsByID(productIds)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	orderID, totalPrice, err := h.createOrder(prodcuts, cart.Items, userID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"total_price": totalPrice,
		"order_id":    orderID,
	})
}

// func (h *Handler) handleUpdateCheckout(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	idStr, ok := vars["id"]

// 	if !ok || strings.TrimSpace(idStr) == "" {
// 		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("missing order ID"))
// 	}

// 	id, err := strconv.Atoi(idStr)
// 	if err != nil {
// 		utils.WriteError(w, http.StatusBadRequest, err)
// 	}

// 	var order = types.Order

// }
