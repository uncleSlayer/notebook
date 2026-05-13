package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"pipeshub"
	"pipeshub/models/components"
)

func NewClient(email, password string) (*pipeshub.SDK, error) {
	baseURL := os.Getenv("PIPESHUB_BASE_URL") + "/api/v1"
	ctx := context.Background()

	s := pipeshub.New(pipeshub.WithServerURL(baseURL))

	initRes, err := s.UserAccount.InitAuth(ctx, components.InitAuthRequest{Email: email})
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
	if authRes.AuthenticateResponse == nil || authRes.AuthenticateResponse.AccessToken == nil {
		return nil, fmt.Errorf("no access token returned")
	}

	return pipeshub.New(
		pipeshub.WithServerURL(baseURL),
		pipeshub.WithSecurity(components.Security{BearerAuth: authRes.AuthenticateResponse.AccessToken}),
	), nil
}
