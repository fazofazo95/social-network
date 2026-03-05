package models

// LoginRequest represents the input for user login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents the output of a successful login
type LoginResponse struct {
	UserID    int
	SessionID string
}

// LoginInput represents login credentials for database queries
type LoginInput struct {
	Email    string
	Password string
}

// // SignUpInput represents signup data for database queries
type SignUpInput struct {
	Username string
	Email    string
	Password string
}
