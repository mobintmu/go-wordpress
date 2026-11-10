package health

import (
	"net/http"

	"go-wordpress/internal/server/drain"

	"github.com/gin-gonic/gin"
)

// Health serves /health and /ready.
type Health struct {
	drain *drain.State
}

func New(d *drain.State) *Health {
	return &Health{drain: d}
}

// Handle reports that the process is running.
func (h *Health) Handle(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{Message: "OK"})
}

// Ready reports whether this instance accepts new requests.
func (h *Health) Ready(c *gin.Context) {
	if h.drain.Draining() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": "SHUTTING_DOWN", "message": "This instance is shutting down."})
		return
	}
	c.JSON(http.StatusOK, HealthResponse{Message: "OK"})
}

type HealthResponse struct {
	Message string `json:"message" example:"OK"`
}
