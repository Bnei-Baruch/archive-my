package authutil

import (
	"github.com/coreos/go-oidc"
	"golang.org/x/net/context"
)

type OIDCTokenVerifier interface {
	Verify(context.Context, string) (*oidc.IDToken, error)
}

type FailoverOIDCTokenVerifier struct {
	verifier *oidc.IDTokenVerifier
}

func (f *FailoverOIDCTokenVerifier) SetVerifier(v *oidc.IDTokenVerifier) {
	f.verifier = v
}

func (f *FailoverOIDCTokenVerifier) Verify(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	return f.verifier.Verify(ctx, rawIDToken)
}
