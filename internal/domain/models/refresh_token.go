package models

type RefreshToken struct {
	ID          *string  `bson:"_id,omitempty"`
	UserGUID    string   `bson:"user_guid"`
	HashedToken [32]byte `bson:"hashed_token"`
	Used        bool     `bson:"used"`
}
