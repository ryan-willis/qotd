package api

import (
	"net/http"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"

	"github.com/gin-gonic/gin"

	"github.com/gorilla/websocket"
	"github.com/ryan-willis/qotd/app"
	"github.com/ryan-willis/qotd/app/room"
	"github.com/ryan-willis/qotd/app/startup"
)

type API struct {
	ctx        *app.Context
	rooms      map[string]*RoomRef
	roomsMutex sync.RWMutex
	upgrader   websocket.Upgrader
}

type ByteChannel chan []byte

type RoomRef struct {
	Room         *room.Room
	ClientsMutex sync.RWMutex
	Clients      map[string]*ByteChannel
	Dead         bool
	Channel      ByteChannel
}

func New() *API {
	ctx := startup.LoadContext()
	return &API{
		ctx:        ctx,
		rooms:      make(map[string]*RoomRef),
		roomsMutex: sync.RWMutex{},
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (a *API) Serve() {
	a.ctx.Log.Info("Server starting...")

	if a.ctx.SentryDSN != "" {
		a.ctx.Log.Info("Initializing Sentry...")
		if err := sentry.Init(sentry.ClientOptions{
			Dsn: a.ctx.SentryDSN,
		}); err != nil {
			a.ctx.Log.Error("Sentry initialization failed!")
		} else {
			a.ctx.Log.Info("Sentry initialized.")
			defer sentry.Recover()
		}
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	if a.ctx.SentryDSN != "" {
		router.Use(sentrygin.New(sentrygin.Options{Repanic: true}))
	}

	// cors is mostly used for development
	// since the ui dev server is on another host
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{},
		AllowMethods:     []string{"POST", "GET"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			if !a.ctx.IsProduction {
				return true
			}
			return origin == "https://qotd.r7.cx"
		},
		MaxAge: 12 * time.Hour,
	}))
	router.Use(static.Serve("/", static.LocalFile("./dist", true)))

	a.addRoutes(router)

	port := "9075"

	a.ctx.Log.Info("Server listening on port " + port)
	if err := router.Run(":" + port); err != nil {
		panic(err)
	}
}

func (a *API) getRoomRef(roomId string) (*RoomRef, bool) {
	a.roomsMutex.RLock()
	roomRef, exists := a.rooms[roomId]
	a.roomsMutex.RUnlock()
	return roomRef, exists
}
