package token

import (
	"MedosTestCase/internal/lib/http/response"
	"MedosTestCase/internal/lib/jwt"
	"MedosTestCase/internal/services"
	"MedosTestCase/internal/services/refresh_token"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	djwt "github.com/golang-jwt/jwt/v5"
	"log/slog"
	"net/http"
)

type Handler struct {
	tokenService *refresh_token.Service
	jwtGen       *jwt.Generator
	log          *slog.Logger
	*chi.Mux
}

func NewHandler(log *slog.Logger, tokenService *refresh_token.Service, jwtGen *jwt.Generator) *Handler {
	handler := Handler{
		log:          log,
		tokenService: tokenService,
		jwtGen:       jwtGen,
		Mux:          chi.NewRouter(),
	}

	handler.Get("/login", handler.login)
	handler.Post("/refresh", handler.refresh)

	return &handler
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.login"
	log := h.log.With(slog.String("operation", op))

	guid := r.URL.Query().Get("guid")
	if guid == "" {
		log.Warn("empty guid")
		response.WriteWithError(w, http.StatusBadRequest, "guid is required")
		return
	}

	accessToken, refreshToken, err := h.jwtGen.GeneratePair(guid)
	if err != nil {
		log.Error("failed to generate tokens", slog.String("error", err.Error()))
		response.WriteWithError(w, http.StatusInternalServerError, "failed to generate tokens")
		return
	}

	err = h.tokenService.Save(guid, refreshToken)
	if err != nil {
		log.Error("failed to save refresh token", slog.String("error", err.Error()))
		response.WriteWithError(w, http.StatusInternalServerError, "failed to generate tokens")
		return
	}

	responseData := struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	response.WriteWithSuccess(w, http.StatusOK, responseData)

	log.Info("tokens generated", slog.String("sub", guid))

	return
}

func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.refresh"
	log := h.log.With(slog.String("operation", op))

	Body := struct {
		RefreshToken string `json:"refresh_token"`
		AccessToken  string `json:"access_token"`
	}{}

	err := json.NewDecoder(r.Body).Decode(&Body)
	if err != nil {
		log.Error("failed to decode request", slog.String("error", err.Error()))
		response.WriteWithError(w, http.StatusBadRequest, "failed to decode request")
		return
	}

	refreshToken := Body.RefreshToken
	accessToken := Body.AccessToken

	if refreshToken == "" {
		log.Warn("empty refresh token")
		response.WriteWithError(w, http.StatusBadRequest, "refresh token is required")
		return
	}

	if accessToken == "" {
		log.Warn("empty access token")
		response.WriteWithError(w, http.StatusBadRequest, "access token is required")
		return
	}

	refreshTokenClaims, err := h.jwtGen.ParseToken(refreshToken)
	if err != nil {
		log.Error("failed to parse refresh token", slog.String("error", err.Error()))
		response.WriteWithError(w, http.StatusBadRequest, "failed to parse token")
		return
	}

	accessTokenClaims, err := h.jwtGen.ParseToken(accessToken)
	if err != nil {
		if !errors.Is(err, djwt.ErrTokenExpired) {
			log.Error("failed to parse access token", slog.String("error", err.Error()))
			response.WriteWithError(w, http.StatusBadRequest, "failed to parse token")
			return
		}
	}

	if err = h.jwtGen.EnsurePair(accessTokenClaims, refreshTokenClaims); err != nil {
		log.Warn("tokens do not match", slog.With("error", err))
		response.WriteWithError(w, http.StatusForbidden, "Invalid credentials")
		return
	}

	sub, err := refreshTokenClaims.Claims.GetSubject()
	if err != nil {
		log.Error("failed to get subject", slog.String("error", err.Error()))
		response.WriteWithError(w, http.StatusInternalServerError, "internal error")
		return
	}

	hashedToken := sha256.Sum256([]byte(refreshToken))
	oldRefreshToken, err := h.tokenService.GetByHash(hashedToken)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			log.Warn("refresh token not found")
			response.WriteWithError(w, http.StatusForbidden, "Invalid credentials")
			return
		}
		log.Error("failed to get refresh token", slog.String("error", err.Error()))
		response.WriteWithError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if oldRefreshToken.UserGUID != sub || oldRefreshToken.Used {
		log.Warn("refresh token mismatch")
		response.WriteWithError(w, http.StatusForbidden, "Invalid credentials")
		return
	}

	newAccessToken, newRefreshToken, err := h.jwtGen.GeneratePair(sub)
	if err != nil {
		log.Error("failed to generate tokens", slog.String("error", err.Error()))
		response.WriteWithError(w, http.StatusInternalServerError, "failed to generate tokens")
		return
	}

	oldRefreshToken.Used = true
	err = h.tokenService.Update(oldRefreshToken)
	if err != nil {
		log.Error("failed to update refresh token", slog.String("error", err.Error()))
		response.WriteWithError(w, http.StatusInternalServerError, "failed to generate tokens")
		return
	}

	err = h.tokenService.Save(sub, newRefreshToken)
	if err != nil {
		log.Error("failed to save refresh token", slog.String("error", err.Error()))
		response.WriteWithError(w, http.StatusInternalServerError, "failed to generate tokens")
		return
	}

	responseData := struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}

	response.WriteWithSuccess(w, http.StatusOK, responseData)

	log.Info("tokens refreshed", slog.String("sub", sub))

	return
}
