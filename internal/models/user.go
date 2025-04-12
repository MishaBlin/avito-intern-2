package models

type User struct {
	ID       string `json:"id,omitempty"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Password string `json:"-"`
}
