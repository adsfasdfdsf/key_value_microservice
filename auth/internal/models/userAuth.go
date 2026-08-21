package models

type UserAuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserAuthResponse struct {
	AccessToken string
}
