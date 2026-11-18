package user

// UpdateProfileInput is a partial update (PATCH): omitted fields are kept,
// an empty bio clears it. Avatar and banner are changed only through the
// upload endpoints, which validate the file.
type UpdateProfileInput struct {
	DisplayName *string `json:"displayName"`
	Bio         *string `json:"bio"`
}

// ChangePasswordInput carries a password change request.
type ChangePasswordInput struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

// ChangeEmailInput carries an email change request.
type ChangeEmailInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
