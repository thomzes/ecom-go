package product

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/thomzes/ecom-go/service/auth"
	"github.com/thomzes/ecom-go/types"
	"github.com/thomzes/ecom-go/utils"
)

type Handler struct {
	store     types.ProductStore
	userStore types.UserStore
}

func NewHandler(store types.ProductStore, userStore types.UserStore) *Handler {
	return &Handler{store: store, userStore: userStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/products", h.handleGetProducts).Methods(http.MethodGet)
	router.HandleFunc("/product/{id}", h.handleGetProduct).Methods(http.MethodGet)
	router.HandleFunc("/products/{ids}", h.handleGetProductsByIds).Methods(http.MethodGet)

	// admin routes
	router.HandleFunc("/product/create", auth.WithJWTAuth(h.HandleCreateProduct, h.userStore)).Methods(http.MethodPost)
	router.HandleFunc("/product/{id}", auth.WithJWTAuth(h.HandleUpdateProduct, h.userStore)).Methods(http.MethodPut)
	router.HandleFunc("/product/{id}", auth.WithJWTAuth(h.HandleDeleteProduct, h.userStore)).Methods(http.MethodDelete)
}

func (h *Handler) handleGetProducts(w http.ResponseWriter, r *http.Request) {
	ps, err := h.store.GetProducts()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, ps)
}

func (h *Handler) handleGetProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	str, ok := vars["id"]

	if !ok {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("missing product ID"))
		return
	}

	id, err := strconv.Atoi(str)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid product ID"))
		return
	}

	product, err := h.store.GetProductByID(id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, product)
}

func (h *Handler) handleGetProductsByIds(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idsStr, ok := vars["ids"]

	if !ok || strings.TrimSpace(idsStr) == "" {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("missing products IDs"))
	}

	parts := strings.Split(idsStr, ",")
	ids := make([]int, 0, len(parts))
	seen := make(map[int]struct{}, len(parts))

	for _, part := range parts {
		p := strings.TrimSpace(part)
		if p == "" {
			continue
		}

		id, err := strconv.Atoi(p)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid product IDs: %q", p))
			return
		}

		if _, dup := seen[id]; dup {
			continue
		}

		seen[id] = struct{}{}

		ids = append(ids, id)
	}

	if len(ids) == 0 {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("no valid product IDs provided"))
		return
	}

	products, err := h.store.GetProductsByID(ids)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, products)

}

func (h *Handler) HandleCreateProduct(w http.ResponseWriter, r *http.Request) {
	// get json payload
	var product types.CreateProductPayload
	if err := utils.ParseJSON(r, &product); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
	}

	// validate the payload
	if err := utils.Validate.Struct(product); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	// validate

	err := h.store.CreateProduct(product)

	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, product)
}

func (h *Handler) HandleUpdateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]

	if !ok || strings.TrimSpace(idStr) == "" {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("missing product ID"))
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid product ID"))
		return
	}

	var product types.Product
	if err := utils.ParseJSON(r, &product); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := utils.Validate.Struct(product); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	if product.ID != 0 && product.ID != id {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("payload ID %d does not match URL ID %d", product.ID, id))
		return
	}

	// update product
	err = h.store.UpdateProduct(product)
	if err != nil {
		if strings.Contains(err.Error(), "not found!") {
			utils.WriteError(w, http.StatusBadRequest, err)
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// get update product to return
	updateProduct, err := h.store.GetProductByID(id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, updateProduct)
}

func (h *Handler) HandleDeleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]

	if !ok || strings.TrimSpace(idStr) == "" {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("missing product ID"))
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid product ID"))
		return
	}

	err = h.store.DeleteProduct(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found!") {
			utils.WriteError(w, http.StatusNotFound, err)
			return
		}
		utils.WriteError(w, http.StatusNotFound, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
