package api

type LoginRequest struct {
	GUID string `form:"guid" binding:"required" validate:"guid"`
}
