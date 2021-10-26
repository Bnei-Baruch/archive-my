package middleware

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc"
	"github.com/gin-gonic/gin"
	pkgerr "github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/Bnei-Baruch/archive-my/databases/mydb/models"
	"github.com/Bnei-Baruch/archive-my/pkg/errs"
)

type Roles struct {
	Roles []string `json:"roles"`
}

type IDTokenClaims struct {
	Acr               string           `json:"acr"`
	AllowedOrigins    []string         `json:"allowed-origins"`
	Aud               interface{}      `json:"aud"`
	AuthTime          int              `json:"auth_time"`
	Azp               string           `json:"azp"`
	Email             string           `json:"email"`
	Exp               int              `json:"exp"`
	FamilyName        string           `json:"family_name"`
	GivenName         string           `json:"given_name"`
	Iat               int              `json:"iat"`
	Iss               string           `json:"iss"`
	Jti               string           `json:"jti"`
	Name              string           `json:"name"`
	Nbf               int              `json:"nbf"`
	Nonce             string           `json:"nonce"`
	PreferredUsername string           `json:"preferred_username"`
	RealmAccess       Roles            `json:"realm_access"`
	ResourceAccess    map[string]Roles `json:"resource_access"`
	SessionState      string           `json:"session_state"`
	Sub               string           `json:"sub"`
	Typ               string           `json:"typ"`

	rolesMap map[string]struct{}
}

func (c *IDTokenClaims) initRoleMap() {
	if c.rolesMap == nil {
		c.rolesMap = make(map[string]struct{})
		if c.RealmAccess.Roles != nil {
			for _, r := range c.RealmAccess.Roles {
				c.rolesMap[r] = struct{}{}
			}
		}
	}
}

func (c *IDTokenClaims) HasAnyRole(roles ...string) bool {
	c.initRoleMap()
	for _, role := range roles {
		if _, ok := c.rolesMap[role]; ok {
			return true
		}
	}
	return false
}

type OIDCTokenVerifier interface {
	Verify(context.Context, string) (*oidc.IDToken, error)
}

type FailoverOIDCTokenVerifier struct {
	verifiers []*oidc.IDTokenVerifier
}

func NewFailoverOIDCTokenVerifier(issuerUrls []string) (OIDCTokenVerifier, error) {
	v := new(FailoverOIDCTokenVerifier)

	for _, url := range issuerUrls {
		provider, err := oidc.NewProvider(context.TODO(), url)
		if err != nil {
			return nil, pkgerr.Wrapf(err, "oidc.NewProvider %s", url)
		}

		v.verifiers = append(v.verifiers, provider.Verifier(&oidc.Config{
			SkipClientIDCheck: true,
		}))
	}

	return v, nil
}

func (v *FailoverOIDCTokenVerifier) Verify(ctx context.Context, tokenStr string) (*oidc.IDToken, error) {
	var token *oidc.IDToken
	var err error

	for _, verifier := range v.verifiers {
		token, err = verifier.Verify(ctx, tokenStr)
		if err == nil {
			return token, nil
		}
	}

	return nil, err
}

func AuthenticationMiddleware(tokenVerifier OIDCTokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := parseToken(c.Request)
		if auth == "" {
			errs.NewUnauthorizedError(pkgerr.Errorf("no `Authorization` header set")).Abort(c)
			return
		}

		token, err := tokenVerifier.Verify(context.TODO(), auth)
		if err != nil {
			errs.NewUnauthorizedError(err).Abort(c)
			return
		}

		var claims IDTokenClaims
		if err := token.Claims(&claims); err != nil {
			errs.NewBadRequestError(pkgerr.Wrap(err, "malformed JWT claims")).Abort(c)
			return
		}
		c.Set("ID_CLAIMS", &claims)

		mydb := c.MustGet("MY_DB").(*sql.DB)
		user, err := getOrCreateUser(mydb, &claims)
		if err != nil {
			errs.NewInternalError(err).Abort(c)
			return
		}
		c.Set("USER", user)

		c.Next()
	}
}

func parseToken(r *http.Request) string {
	authHeader := strings.Split(strings.TrimSpace(r.Header.Get("Authorization")), " ")
	if len(authHeader) == 2 &&
		strings.ToLower(authHeader[0]) == "bearer" &&
		len(authHeader[1]) > 0 {
		return authHeader[1]
	}
	return ""
}

func getOrCreateUser(exec boil.Executor, claims *IDTokenClaims) (*models.User, error) {
	//TODO: wrap in transaction
	user, err := models.Users(models.UserWhere.AccountsID.EQ(claims.Sub)).One(exec)
	if err == nil {
		return user, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, pkgerr.Wrap(err, "lookup user in DB")
	}

	user = &models.User{
		AccountsID: claims.Sub,
		Email:      null.StringFrom(claims.Email),
		FirstName:  null.StringFrom(claims.GivenName),
		LastName:   null.StringFrom(claims.FamilyName),
		Disabled:   false,
	}
	if err := user.Insert(exec, boil.Infer()); err != nil {
		return nil, pkgerr.Wrap(err, "create new user in DB")
	}
	return user, nil
}
