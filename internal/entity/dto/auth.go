package dto

type AuthCredentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserRegister struct {
	AuthCredentials
	Name    string `json:"first_name" valid:"runelength(2|30)"`
	Surname string `json:"last_name" valid:"runelength(2|30)"`
}

type Login struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type AuthResponse struct {
	UserId int `json:"user_id"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type LoginExistsRequest struct {
	Login string `json:"login"`
}

type LoginExistsResponse struct {
	Exists bool `json:"exists"`
}
