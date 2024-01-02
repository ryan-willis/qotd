package api

import (
	"net/http"
	"sync"
	"time"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/gin-gonic/gin"
	"github.com/ryan-willis/qotd/app/logger"
	"github.com/ryan-willis/qotd/app/room"
	"github.com/ryan-willis/qotd/app/util"
)

func keyFunc(c *gin.Context) string {
	return c.ClientIP()
}

func errorHandler(c *gin.Context, info ratelimit.Info) {
	c.JSON(http.StatusTooManyRequests, gin.H{
		"error_code":    "TOO_MANY_REQUESTS",
		"error_message": "Too many requests. Try again in " + time.Until(info.ResetTime).Round(time.Second).String(),
	})
}

func (a *API) addRoutes(e *gin.Engine) {
	// prevent room creation spam
	createRoomStore := ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
		Rate:  time.Minute,
		Limit: 2, // 2 rooms created per minute
	})
	createRoomRateLimitMiddleware := ratelimit.RateLimiter(createRoomStore, &ratelimit.Options{
		ErrorHandler: errorHandler,
		KeyFunc:      keyFunc,
	})
	e.POST("/rooms", createRoomRateLimitMiddleware, a.CreateRoomRoute())

	// prevent room fetching spam
	getRoomStore := ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
		Rate:  time.Minute,
		Limit: 6,
	})
	getRoomRateLimitMiddleware := ratelimit.RateLimiter(getRoomStore, &ratelimit.Options{
		ErrorHandler: errorHandler,
		KeyFunc:      keyFunc,
	})
	e.GET("/rooms/:roomId", getRoomRateLimitMiddleware, a.GetRoomRoute())

	// TOOD: rate limit websocket connections
	e.GET("/ws", a.GetRoomWebsocket())

	if !a.ctx.IsProduction {
		// expose stats route for debugging
		e.GET("/stats", a.Stats())
	}
}

func (a *API) CreateRoomRoute() gin.HandlerFunc {
	return func(c *gin.Context) {
		room, err := room.New()
		if err != nil {
			a.ctx.Log.Error("error creating room", logger.Field("err", err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error_code":    "INTERNAL_SERVER_ERROR",
				"error_message": "could not create room",
			})
			return
		}
		clients := make(map[string]*ByteChannel)
		roomRef := &RoomRef{
			Dead:         false,
			Channel:      make(ByteChannel),
			Room:         room,
			Clients:      clients,
			ClientsMutex: sync.RWMutex{},
		}
		a.roomsMutex.Lock()
		a.rooms[room.ID] = roomRef
		a.roomsMutex.Unlock()
		go a.roomHandler(roomRef)
		a.ctx.Log.Info("room created", logger.Field("roomId", room.ID))
		c.JSON(http.StatusCreated, gin.H{
			"room": room,
		})
	}
}

func (a *API) GetRoomRoute() gin.HandlerFunc {
	return func(c *gin.Context) {
		roomId := c.Param("roomId")
		if !util.IsValidRoomCode(roomId) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error_code":    "INVALID_ROOM_CODE",
				"error_message": "room code is invalid",
			})
			return
		}
		roomRef, exists := a.getRoomRef(roomId)
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{
				"error_code":    "ROOM_NOT_FOUND",
				"error_message": "room not found",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"room": roomRef,
		})
	}
}
