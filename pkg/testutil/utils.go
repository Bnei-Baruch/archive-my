package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/coreos/go-oidc"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
)

const KEYCKLOAK_ID = "test_keycloak_user_id"

type OIDCTokenVerifier struct{}

func (v *OIDCTokenVerifier) Verify(context.Context, string) (*oidc.IDToken, error) {
	return &oidc.IDToken{Subject: KEYCKLOAK_ID}, nil
}

func PrepareContext(d interface{}) (*gin.Context, *httptest.ResponseRecorder, error) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("KC_ID", KEYCKLOAK_ID)
	jsonbytes, err := json.Marshal(d)
	if err != nil {
		return nil, nil, err
	}
	c.Request = &http.Request{
		Header: make(http.Header),
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(jsonbytes))
	return c, w, nil
}
