package server

import (
	"MedosTestCase/internal/lib/jwt"
	"MedosTestCase/internal/server/handlers/token"
	"MedosTestCase/internal/services/refresh_token"
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type Server struct {
	Address             string `yaml:"address" env-required:"true"`
	RefreshTokenService *refresh_token.Service
	Log                 *slog.Logger
	Srv                 *http.Server
	JwtGen              *jwt.Generator
}

func NewServer(address string, log *slog.Logger, refreshTokenService *refresh_token.Service, jwtGen *jwt.Generator) *Server {
	return &Server{
		Address:             address,
		Log:                 log,
		RefreshTokenService: refreshTokenService,
		JwtGen:              jwtGen,
	}
}

func (s *Server) Run() {
	s.Log.Info("Server is running", slog.String("address", s.Address))
	s.Srv = &http.Server{
		Addr:    s.Address,
		Handler: nil,
	}

	router := chi.NewRouter()

	tokensHandler := token.NewHandler(s.Log, s.RefreshTokenService, s.JwtGen)
	router.Mount("/tokens", tokensHandler)

	s.Srv.Handler = router

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
		s.Log.Info("Gratefully shutting down server")
		if err := s.Srv.Shutdown(context.Background()); err != nil {
			s.Log.Error("Failed to shutdown server", slog.String("error", err.Error()))
		}
	}()

	if err := s.Srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.Log.Error("Failed to listen and serve", slog.String("error", err.Error()))
	}

	s.Log.Info("Server stopped")
}
