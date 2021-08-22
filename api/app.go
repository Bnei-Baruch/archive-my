package api

import (
	"context"
	"database/sql"

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

	// cors
	opt := cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PATCH", "HEAD", "OPTIONS", "DELETE"},
		AllowedHeaders: []string{"Origin", "Authorization", "Accept", "Content-Type", "X-Requested-With", "X-Request-ID"},
	}
	router.Use(cors.New(opt))
	//router.Use(cors.Default())

	router.GET("/health_check", a.HealthCheckHandler)
	router.GET("/like_count", a.handleLikeCount)

	rest := router.Group("/rest")
	rest.Use(authutil.AuthenticationMiddleware(a.Verifier))

	rest.GET("/playlists", a.handleGetPlaylists)
	rest.POST("/playlists", a.handleCreatePlaylist)
	rest.DELETE("/playlists", a.handleDeletePlaylist)
	rest.GET("/playlists/:id", a.handleGetPlaylist)
	rest.PATCH("/playlists/:id", a.handleUpdatePlaylist)
	rest.POST("/playlists/:id", a.handleAddToPlaylist)
	rest.GET("/playlist_items", a.handleGetPlaylistItems)
	rest.DELETE("/playlist_items/:id", a.handleDeleteFromPlaylist)
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
