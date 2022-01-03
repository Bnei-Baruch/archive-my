package middleware

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"sync"

	"github.com/coreos/go-oidc"
	"github.com/gin-gonic/gin"
	pkgerr "github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/Bnei-Baruch/archive-my/databases/mydb/models"
	"github.com/Bnei-Baruch/archive-my/pkg/errs"
	"github.com/Bnei-Baruch/archive-my/pkg/utils"
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

type Auth struct {
	mu     sync.Mutex
	claims *IDTokenClaims
	db     *sql.DB
	user   *models.User
}

func (a *Auth) AuthenticationMiddleware(tokenVerifier OIDCTokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := a.parseToken(c.Request)
		if auth == "" {
			errs.NewUnauthorizedError(pkgerr.Errorf("no `Authorization` header set")).Abort(c)
			return
		}

		token, err := tokenVerifier.Verify(context.TODO(), auth)
		if err != nil {
			errs.NewUnauthorizedError(err).Abort(c)
			return
		}
		a.claims = new(IDTokenClaims)
		if err := token.Claims(a.claims); err != nil {
			errs.NewBadRequestError(pkgerr.Wrap(err, "malformed JWT claims")).Abort(c)
			return
		}
		c.Set("ID_CLAIMS", a.claims)

		a.db = c.MustGet("MY_DB").(*sql.DB)
		err = a.getOrCreateUser()
		if err != nil {
			errs.NewInternalError(err).Abort(c)
			return
		}
		c.Set("USER", a.user)

		c.Next()
	}
}

func (a *Auth) CheckModeratorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if ok := a.claims.HasAnyRole("kmedia_moderator"); !ok {
			errs.NewForbiddenError(nil).Abort(c)
			return
		}

		c.Next()
	}
}

func (a *Auth) getOrCreateUser() error {
	if err := a.fetchUserFromDB(); err != nil || a.user != nil {
		return err
	}

	user := &models.User{
		AccountsID: a.claims.Sub,
		Email:      null.StringFrom(a.claims.Email),
		FirstName:  null.StringFrom(a.claims.GivenName),
		LastName:   null.StringFrom(a.claims.FamilyName),
		Disabled:   false,
	}

	tx, err := a.db.Begin()
	utils.Must(err)
	// check if not unique on DB - "23505": "unique_violation",
	errDB := user.Insert(tx, boil.Infer())
	if errDB != nil {
		utils.Must(tx.Rollback())
		if !strings.Contains(errDB.Error(), "pq: duplicate key value violates unique constraint \"users_accounts_id_key\"") {
			return pkgerr.Wrap(errDB, "create new user in DB")
		}
	} else {
		utils.Must(tx.Commit())
	}
	return a.fetchUserFromDB()
}

func (a *Auth) fetchUserFromDB() error {
	var err error
	a.user, err = models.Users(models.UserWhere.AccountsID.EQ(a.claims.Sub)).One(a.db)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return pkgerr.Wrap(err, "lookup user in DB")
	}
	return nil
}

func (a *Auth) parseToken(r *http.Request) string {
	authHeader := strings.Split(strings.TrimSpace(r.Header.Get("Authorization")), " ")
	if len(authHeader) == 2 &&
		strings.ToLower(authHeader[0]) == "bearer" &&
		len(authHeader[1]) > 0 {
		return authHeader[1]
	}
	return ""
}
