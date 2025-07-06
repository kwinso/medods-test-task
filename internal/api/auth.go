package api

type LoginRequest struct {
	// GUID for the user that is logging in
	GUID string `json:"guid" binding:"required" validate:"guid" example:"12345678-1234-1234-1234-123456789012"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"base64-encoded-token"`
}

type TokenPair struct {
	// AccessToken is a JWT token that can be used to access the API
	AccessToken string `json:"access_token" `
	// RefreshToken is a randomly generated base64 string that can be used to refresh the access token
	// It is valid for 30 days
	// Refresh token can only be used to refresh a single access token it was issued with.
	// After refreshing, the refresh token is no longer valid and cannot be used again.
	RefreshToken string `json:"refresh_token"`
}

// GetMeResponse holds a response for the /me route
// @Description	Contains the GUID for the authenticated user
type GetMeResponse struct {
	Guid string `json:"guid" example:"12345678-1234-1234-1234-123456789012"`
}
