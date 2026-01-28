package models

// LoginRequest represents the input for user login
type LoginRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SignUpRequest represents the input for user signup
type SignUpRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents the output of a successful login
type LoginResponse struct {
	UserID    int
	SessionID string
}
