package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	cors "github.com/rs/cors/wrapper/gin"
	"github.com/rs/zerolog/log"

	"github.com/Bnei-Baruch/archive-my/common"
	"github.com/Bnei-Baruch/archive-my/instrumentation"
	"github.com/Bnei-Baruch/archive-my/lib/chronicles"
	"github.com/Bnei-Baruch/archive-my/middleware"
	"github.com/Bnei-Baruch/archive-my/pkg/utils"
)

type App struct {
	Router     *gin.Engine
	DB         *sql.DB
	chronicles *chronicles.Chronicles
}

func (a *App) Initialize() {
	log.Info().Msg("initializing app")

	log.Info().Msg("Initializing token verifier")
	verifier, err := middleware.NewFailoverOIDCTokenVerifier(common.Config.AccountsUrls)
	utils.Must(err)

	log.Info().Msg("Setting up connection to MyDB")
	db, err := sql.Open("postgres", common.Config.MyDBUrl)
	utils.Must(err)

	a.InitializeWithDeps(db, verifier)
}

func (a *App) InitializeWithDeps(db *sql.DB, tokenVerifier middleware.OIDCTokenVerifier) {
	a.DB = db

	gin.SetMode(common.Config.GinMode)
	a.Router = gin.New()

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowedHeaders: []string{"Origin", "Accept", "Content-Type", "Authorization", "X-Requested-With", "X-Request-ID"},
		MaxAge:         3600,
	})

	a.Router.Use(
		middleware.LoggingMiddleware(),
		middleware.RecoveryMiddleware(),
		middleware.ErrorHandlingMiddleware(),
		corsMiddleware,
		middleware.DataStoresMiddleware(a.DB))

	a.initRoutes(tokenVerifier)

	a.chronicles = new(chronicles.Chronicles)
	a.chronicles.Init()
	instrumentation.Stats.Init()
}

func (a *App) Run() {
	defer a.Shutdown()

	a.chronicles.Run()
	addr := common.Config.ListenAddress
	log.Info().Msgf("app run %s", addr)
	if err := a.Router.Run(addr); err != nil {
		log.Fatal().Err(err).Msg("Router.Run")
	}
}

func (a *App) Shutdown() {
	a.chronicles.Stop()

	if err := a.DB.Close(); err != nil {
		log.Error().Err(err).Msg("DB.close")
	}
}

func (a *App) initRoutes(verifier middleware.OIDCTokenVerifier) {
	a.Router.GET("/health_check", a.HealthCheckHandler)
	a.Router.GET("/metrics", a.MakePrometheusHandler())

	a.Router.GET("/reaction_count", a.handleReactionCount)
	// TODO: public endpoint for public playlists (get by UID) string all internal IDs

	rest := a.Router.Group("/rest")
	auth := middleware.Auth{}
	rest.Use(auth.AuthenticationMiddleware(verifier))

	rest.GET("/playlists", a.handleGetPlaylists)
	rest.POST("/playlists", a.handleCreatePlaylist)
	rest.GET("/playlists/:id", a.handleGetPlaylist)
	rest.PUT("/playlists/:id", a.handleUpdatePlaylist)
	rest.DELETE("/playlists/:id", a.handleDeletePlaylist)

	// These would have been more "restful" had gin router support it ...
	rest.POST("/playlists/:id/add_items", a.handleAddPlaylistItems)
	rest.PUT("/playlists/:id/update_items", a.handleUpdatePlaylistItems)
	rest.DELETE("/playlists/:id/remove_items", a.handleRemovePlaylistItems)
	// end gin router rants

	rest.GET("/reactions", a.handleGetReactions)
	rest.POST("/reactions", a.handleAddReactions)
	rest.DELETE("/reactions", a.handleRemoveReactions)
	rest.GET("/subscriptions", a.handleGetSubscriptions)
	rest.POST("/subscriptions", a.handleSubscribe)
	rest.DELETE("/subscriptions/:id", a.handleUnsubscribe)
	rest.GET("/history", a.handleGetHistory)
	rest.DELETE("/history/:id", a.handleDeleteHistory)
	rest.GET("/bookmarks", a.handleGetBookmarks)
	rest.POST("/bookmarks", a.handleCreateBookmark)
	rest.PUT("/bookmarks/:id", a.handleUpdateBookmark)
	rest.DELETE("/bookmarks/:id", a.handleDeleteBookmark)
	/*
		rest.GET("/folders", a.handleGetFolders)
		rest.POST("/folders", a.handleCreateFolder)
		rest.PUT("/folders/:id", a.handleUpdateFolder)
		rest.DELETE("/folders/:id", a.handleDeleteFolder)
	*/
}
