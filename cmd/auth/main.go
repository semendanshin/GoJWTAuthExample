package main

import (
	"MedosTestCase/config"
	"MedosTestCase/internal/lib/jwt"
	refreshTokenMongo "MedosTestCase/internal/repositories/refresh_token/mongo"
	"MedosTestCase/internal/server"
	"MedosTestCase/internal/services/refresh_token"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	defaultLogger "log"
	"log/slog"
	"os"
)

const (
	Development = "dev"
	Production  = "prod"
	Test        = "test"
)

func main() {
	// Read config
	cfgPath := config.FetchPath()
	defaultLogger.Println("Using config path: ", cfgPath)
	cfg := config.MustParseConfig(cfgPath)

	// Init logger
	log := InitLogger(cfg.Env)
	log.Info("Logger initialized", slog.Any("env", cfg.Env))

	// Init database
	connStr := fmt.Sprintf("mongodb://%s:%s@%s:%d", cfg.Mongo.User, cfg.Mongo.Pass, cfg.Mongo.Host, cfg.Mongo.Port)
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(connStr))

	if err != nil {
		log.Error("Failed to connect to database", slog.With("error", err.Error()))
		return
	}

	if err = client.Ping(context.Background(), nil); err != nil {
		log.Error("Failed to ping database", slog.With("error", err.Error()))
		return
	}

	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			log.Error("Failed to disconnect from database", slog.With("error", err.Error()))
		}
	}()

	// Init repositories
	refreshTokenRepo := refreshTokenMongo.NewRefreshTokenRepository(
		client.Database(cfg.Mongo.Name).Collection("refresh_tokens"),
		log,
	)

	// Init services
	refreshTokenService := refresh_token.NewService(refreshTokenRepo, log)

	// Init JWTGenerator
	jwtGenerator := jwt.NewGenerator(cfg.Tokens.Secret, cfg.Tokens.AccessTTL, cfg.Tokens.RefreshTTL)

	// Init server
	srv := server.NewServer(
		cfg.Server.Address,
		log,
		refreshTokenService,
		jwtGenerator,
	)

	// Start server
	srv.Run()
}

func InitLogger(env string) *slog.Logger {
	var log *slog.Logger
	if env == Production {
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
		return log
	} else if env == Test {
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
		return log
	} else {
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
		return log
	}
}
