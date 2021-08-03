package api

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/coreos/go-oidc"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	cors "github.com/rs/cors/wrapper/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"archive-my/pkg/authutil"
	"archive-my/pkg/utils"
)

type App struct {
	DB       *sql.DB
	Router   *gin.Engine
	Verifier authutil.OIDCTokenVerifier
}

func (a *App) SetVerifier(verifier authutil.OIDCTokenVerifier) {
	a.Verifier = verifier
}

func (a *App) SetDB(db *sql.DB) {
	a.DB = db
}

func (a *App) InitDeps() {
	log.Info("Initializing token verifier")
	provider, err := oidc.NewProvider(context.TODO(), viper.GetString("app.issuer"))
	utils.Must(err)
	v := &authutil.FailoverOIDCTokenVerifier{}
	v.SetVerifier(provider.Verifier(&oidc.Config{SkipClientIDCheck: true}))
	a.Verifier = v

	log.Info("Setting up connection to DB")
	a.DB, err = sql.Open("postgres", viper.GetString("app.mydb"))
	utils.Must(err)
}

func (a *App) Run() {
	defer func() {
		log.Info("close connection to My DB")
		utils.Must(a.DB.Close())
	}()

	utils.Must(a.DB.Ping())
	boil.DebugMode = viper.GetString("server.boiler-mode") == "debug"
	boil.SetDB(a.DB)

	utils.Must(a.setupRoutes())
}

func (a *App) setupRoutes() error {
	// Setup gin
	gin.SetMode(viper.GetString("server.mode"))
	router := gin.Default()
	router.GET("/health_check", a.HealthCheckHandler)

	// cors
	opt := cors.Options{}
	opt.AllowedHeaders = append(opt.AllowedHeaders, "Authorization", "X-Request-ID")
	opt.AllowedMethods = append(opt.AllowedMethods, http.MethodDelete)
	cors.Default()
	router.Use(cors.New(opt))

	rest := router.Group("/rest", authutil.AuthenticationMiddleware(a.Verifier))
	rest.GET("/playlists", a.handleGetPlaylists)
	rest.POST("/playlists", a.handleCreatePlaylist)
	rest.PATCH("/playlists/:id", a.handleUpdatePlaylist)
	rest.DELETE("/playlists/:id", a.handleDeletePlaylist)
	rest.POST("/playlists/:id/units", a.handleAddToPlaylist)
	rest.DELETE("/playlists/:id/units", a.handleDeleteFromPlaylist)
	rest.GET("/likes", a.handleGetLikes)
	rest.POST("/likes", a.handleAddLikes)
	rest.DELETE("/likes", a.handleRemoveLikes)
	rest.GET("/subscriptions", a.handleGetSubscriptions)
	rest.POST("/subscriptions", a.handleSubscribe)
	rest.DELETE("/subscriptions", a.handleUnsubscribe)
	rest.GET("/history", a.handleGetHistory)
	rest.DELETE("/history", a.handleDeleteHistory)

	return router.Run(viper.GetString("server.bind-address"))
}
