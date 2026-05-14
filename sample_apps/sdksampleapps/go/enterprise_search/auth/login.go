package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"undefined"
	"undefined/models/components"
)

func NewClient(email, password string) (*undefined.SDK, error) {
	baseURL := os.Getenv("PIPESHUB_BASE_URL") + "/api/v1"
	ctx := context.Background()

	s := undefined.New(undefined.WithServerURL(baseURL))

	initRes, err := s.UserAccount.InitAuth(ctx, &components.InitAuthRequest{Email: &email})
	if err != nil {
		return nil, fmt.Errorf("init auth: %w", err)
	}
	sessionToken := http.Header(initRes.Headers).Get("x-session-token")

	authRes, err := s.UserAccount.Authenticate(ctx, sessionToken, components.AuthenticateRequest{
		Method: components.MethodPassword,
		Credentials: components.CreateCredentialsPasswordCredentials(
			components.PasswordCredentials{Password: password},
		),
	})
	if err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}
	if authRes.AuthenticateResponse == nil || authRes.AuthenticateResponse.AuthenticateFinalResponse == nil {
		return nil, fmt.Errorf("authenticate: expected final response, got multi-step or empty")
	}
	accessToken := authRes.AuthenticateResponse.AuthenticateFinalResponse.AccessToken

	return undefined.New(
		undefined.WithServerURL(baseURL),
		undefined.WithSecurity(components.Security{BearerAuth: &accessToken}),
	), nil
}
