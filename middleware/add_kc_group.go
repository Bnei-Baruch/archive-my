package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/Bnei-Baruch/archive-my/common"
	"github.com/Bnei-Baruch/archive-my/databases/mydb/models"
)

func AddToKeycloakGroup() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := c.MustGet("USER").(*models.User)
		claims := c.MustGet("ID_CLAIMS").(*IDTokenClaims)

		addKC := len(claims.RealmAccess.Roles) == 0
		for _, role := range claims.RealmAccess.Roles {
			if role == common.Config.KmediaKCRole {
				return
			}
			if role == common.Config.NewUserKCRole {
				addKC = true
			}
		}

		if addKC {
			sendAddKCGroup(user.AccountsID)
		}
	}
}

func sendAddKCGroup(userId string) {
	url := fmt.Sprintf("%s&user_id=%s", common.Config.KCGroupUrl, userId)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		log.Error().Err(err).Msgf("Error on send to KC")
	}

	if resp == nil || resp.StatusCode != http.StatusOK {
		log.Error().Msgf("Error on add to KC group")
	}
}
