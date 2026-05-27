package configuration

import (
	"log"
	"strings"

	"github.com/gin-gonic/gin"
)

func ConfigureLogger(appEnv string) {
	log.SetFlags(log.LstdFlags | log.LUTC | log.Lshortfile)
	if strings.EqualFold(appEnv, "production") {
		gin.SetMode(gin.ReleaseMode)
		return
	}
	gin.SetMode(gin.DebugMode)
}
