package repository

import (
	"golang.org/x/oauth2"
	oauth2v2 "google.golang.org/api/oauth2/v2"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type GoogleAuthToken struct {
	gorm.Model
	UserInfo  datatypes.JSONType[*oauth2v2.Userinfo] `json:"user_info"`
	AuthToken datatypes.JSONType[*oauth2.Token]      `json:"auth_token"`
}

func GetGoogleAuthToken(db *gorm.DB) *GoogleAuthToken {
	var token GoogleAuthToken

	if err := db.First(&token).Error; err != nil {
		return nil
	}

	return &token
}

func SaveGoogleAuthToken(db *gorm.DB, userInfo *oauth2v2.Userinfo, authToken *oauth2.Token) *GoogleAuthToken {
	gToken := GetGoogleAuthToken(db)
	if gToken == nil {
		gToken = &GoogleAuthToken{}
	}

	gToken.AuthToken = datatypes.NewJSONType(authToken)
	gToken.UserInfo = datatypes.NewJSONType(userInfo)
	db.Save(gToken)

	return gToken
}

func DeleteGoogleAuthToken(db *gorm.DB) {
	var token GoogleAuthToken

	db.First(&token)
	db.Delete(&token)
}
