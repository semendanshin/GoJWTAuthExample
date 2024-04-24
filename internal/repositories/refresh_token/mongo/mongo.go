package mongo

import (
	"MedosTestCase/internal/domain/models"
	"MedosTestCase/internal/services"
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log/slog"
)

type RefreshTokenRepository struct {
	collection *mongo.Collection
	log        *slog.Logger
}

func NewRefreshTokenRepository(collection *mongo.Collection, log *slog.Logger) *RefreshTokenRepository {
	return &RefreshTokenRepository{
		collection: collection,
		log:        log,
	}
}

func (r *RefreshTokenRepository) CreateRefreshToken(refreshToken *models.RefreshToken) error {
	_, err := r.collection.InsertOne(context.Background(), refreshToken)
	if err != nil {
		if !mongo.IsDuplicateKeyError(err) {
			return err
		}
		return services.ErrAlreadyExists
	}

	return nil
}

func (r *RefreshTokenRepository) GetRefreshTokenByUserGUID(guid string) ([]*models.RefreshToken, error) {
	cur, err := r.collection.Find(context.Background(), map[string]string{"user_guid": guid})
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, services.ErrNotFound
		}
		return nil, err
	}

	var refreshTokens []*models.RefreshToken
	for cur.Next(context.Background()) {
		var token models.RefreshToken
		err := cur.Decode(&token)
		if err != nil {
			return nil, err
		}
		refreshTokens = append(refreshTokens, &token)
	}

	return refreshTokens, nil
}

func (r *RefreshTokenRepository) UpdateRefreshToken(refreshToken *models.RefreshToken) error {
	objId, err := primitive.ObjectIDFromHex(*refreshToken.ID)
	if err != nil {
		return err
	}
	tokenCopy := *refreshToken
	tokenCopy.ID = nil
	filter := bson.D{{"_id", objId}}
	set := bson.D{{"$set", tokenCopy}}
	_, err = r.collection.UpdateOne(context.Background(), filter, set)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return services.ErrNotFound
		}
	}
	return nil
}

func (r *RefreshTokenRepository) GetRefreshTokenByHash(hash [32]byte) (*models.RefreshToken, error) {
	refreshToken := &models.RefreshToken{}
	err := r.collection.FindOne(context.Background(), map[string][32]byte{"hashed_token": hash}).Decode(refreshToken)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, services.ErrNotFound
		}
		return nil, err
	}

	return refreshToken, nil
}
