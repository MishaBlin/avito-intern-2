package response

type UserResponse struct {
	ID    string `json:"id,omitempty"`
	Email string `json:"email"`
	Role  string `json:"role"`
}
