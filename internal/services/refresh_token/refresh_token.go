package refresh_token

import (
	"MedosTestCase/internal/domain/models"
	"crypto/sha256"
	"log/slog"
)

type Repository interface {
	CreateRefreshToken(refreshToken *models.RefreshToken) error
	GetRefreshTokenByUserGUID(guid string) ([]*models.RefreshToken, error)
	UpdateRefreshToken(refreshToken *models.RefreshToken) error
	GetRefreshTokenByHash(hash [32]byte) (*models.RefreshToken, error)
}

type Service struct {
	repo Repository
	log  *slog.Logger
}

func NewService(repo Repository, log *slog.Logger) *Service {
	return &Service{
		repo: repo,
		log:  log,
	}
}

func (s *Service) Save(guid string, refreshToken string) error {
	hashedToken := sha256.Sum256([]byte(refreshToken))
	refreshTokenModel := &models.RefreshToken{
		UserGUID:    guid,
		HashedToken: hashedToken,
		Used:        false,
	}

	return s.repo.CreateRefreshToken(refreshTokenModel)
}

func (s *Service) Get(guid string) ([]*models.RefreshToken, error) {
	return s.repo.GetRefreshTokenByUserGUID(guid)
}

func (s *Service) Update(refreshToken *models.RefreshToken) error {
	return s.repo.UpdateRefreshToken(refreshToken)
}

func (s *Service) GetByHash(hash [32]byte) (*models.RefreshToken, error) {
	return s.repo.GetRefreshTokenByHash(hash)
}
