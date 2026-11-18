package user

// UpdateProfileInput carries editable profile fields.
type UpdateProfileInput struct {
	DisplayName string  `json:"displayName"`
	Bio         *string `json:"bio"`
	AvatarURL   *string `json:"avatarUrl"`
	BannerURL   *string `json:"bannerUrl"`
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
