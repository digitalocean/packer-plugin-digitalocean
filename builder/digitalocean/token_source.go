package digitalocean

import (
	"golang.org/x/oauth2"
)

type APITokenSource struct {
	AccessToken string
}

func (t *APITokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{
		AccessToken: t.AccessToken,
	}, nil
}
