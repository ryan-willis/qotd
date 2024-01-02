package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *API) Stats() gin.HandlerFunc {
	return func(c *gin.Context) {
		var roomList []string
		var clients int = 0
		a.roomsMutex.RLock()
		for roomId := range a.rooms {
			roomList = append(roomList, roomId)
			a.rooms[roomId].ClientsMutex.RLock()
			for range a.rooms[roomId].Clients {
				clients++
			}
			a.rooms[roomId].ClientsMutex.RUnlock()
		}
		a.roomsMutex.RUnlock()
		c.JSON(http.StatusOK, gin.H{
			"rooms":   roomList,
			"clients": clients,
		})
	}
}
