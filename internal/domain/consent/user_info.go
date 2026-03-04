package consent

// UserInfo represents the current authenticated user's information.
// This is used for the consent management UI to display user profile data.
type UserInfo struct {
	// Principal is the user's unique identifier (email, UUID, etc.)
	Principal string `json:"principal"`

	// DisplayName is a human-readable name for the user
	DisplayName string `json:"displayName"`

	// Email is an optional email address extracted from JWT claims.
	// Available when JWT pre-authentication is configured with an email CEL expression.
	// Nil when using plain header pre-auth or when the JWT does not contain an email claim.
	Email *string `json:"email,omitempty"`

	// PictureURL is an optional URL to the user's profile picture/avatar
	PictureURL *string `json:"pictureUrl,omitempty"`
}
