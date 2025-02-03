package repository

import (
	"golang.org/x/oauth2"
	oauth2v2 "google.golang.org/api/oauth2/v2"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// UNSYNCED MODEL

type GoogleAuthToken struct {
	gorm.Model
	UserInfo  datatypes.JSONType[*oauth2v2.Userinfo] `json:"user_info"`
	AuthToken datatypes.JSONType[*oauth2.Token]      `json:"auth_token"`
}

type GoogleAuthTokenRepository struct {
	BaseRepository
}

func NewGoogleAuthTokenRepository(unSyncedDB *gorm.DB) *GoogleAuthTokenRepository {
	return &GoogleAuthTokenRepository{
		BaseRepository: BaseRepository{db: unSyncedDB},
	}
}

func (r *GoogleAuthTokenRepository) GetGoogleAuthToken() *GoogleAuthToken {
	var token GoogleAuthToken

	if err := r.db.First(&token).Error; err != nil {
		return nil
	}

	return &token
}

func (r *GoogleAuthTokenRepository) SaveGoogleAuthToken(userInfo *oauth2v2.Userinfo, authToken *oauth2.Token) *GoogleAuthToken {
	gToken := r.GetGoogleAuthToken()
	if gToken == nil {
		gToken = &GoogleAuthToken{}
	}

	gToken.AuthToken = datatypes.NewJSONType(authToken)
	gToken.UserInfo = datatypes.NewJSONType(userInfo)
	r.db.Save(gToken)

	return gToken
}

func (r *GoogleAuthTokenRepository) UpdateGoogleAuthToken(token *oauth2.Token) {
	gToken := r.GetGoogleAuthToken()
	if gToken != nil {
		gToken.AuthToken = datatypes.NewJSONType(token)
		r.db.Save(gToken)
	}

}

func (r *GoogleAuthTokenRepository) DeleteGoogleAuthToken() {
	var token GoogleAuthToken

	r.db.First(&token)
	r.db.Delete(&token)
}
