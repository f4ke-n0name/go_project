package handler

import (
	"encoding/json"
	"net/http"

	auth "github.com/f4ke-n0name/autoparts-hub/gen/auth"
	"github.com/f4ke-n0name/autoparts-hub/services/api-gateway/internal/utils"
)

type AuthHandler struct {
	client auth.AuthClient
}

func NewAuthHandler(client auth.AuthClient) *AuthHandler {
	return &AuthHandler{client: client}
}

// Register godoc
// @Summary      Register a new user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body RegisterRequest true "Register request"
// @Success      200 {object} TokenResponse
// @Failure      400 {object} map[string]string
// @Router       /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.Register(r.Context(), &auth.RegisterRequest{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  req.Password,
	})
	if err != nil {
		utils.WriteGRPCError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, TokenResponse{
		AccessToken:  resp.GetAccessToken(),
		RefreshToken: resp.GetRefreshToken(),
	})
}

// Login godoc
// @Summary      Login
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "Login request"
// @Success      200 {object} TokenResponse
// @Failure      400 {object} map[string]string
// @Router       /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.Login(r.Context(), &auth.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		utils.WriteGRPCError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, TokenResponse{
		AccessToken:  resp.GetAccessToken(),
		RefreshToken: resp.GetRefreshToken(),
	})
}

// RefreshToken godoc
// @Summary      Refresh access token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body RefreshRequest true "Refresh request"
// @Success      200 {object} TokenResponse
// @Failure      400 {object} map[string]string
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.RefreshToken(r.Context(), &auth.RefreshRequest{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		utils.WriteGRPCError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, TokenResponse{
		AccessToken:  resp.GetAccessToken(),
		RefreshToken: resp.GetRefreshToken(),
	})
}

type RegisterRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
