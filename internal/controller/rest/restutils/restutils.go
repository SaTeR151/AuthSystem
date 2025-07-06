package restutils

import (
	"encoding/base64"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Error(c *gin.Context, err string, code int) {
	c.AbortWithStatusJSON(code, gin.H{"Message": err})
	return
}

func SetCookieTokens(c *gin.Context, accessT string, refreshT string) {
	rtB64 := base64.StdEncoding.EncodeToString([]byte(refreshT))
	timeExp, err := strconv.Atoi(os.Getenv("COOKIEEXPIRES"))
	if err != nil {
		logrus.Error(err)
		Error(c, err.Error(), http.StatusInternalServerError)
		return
	}
	c.SetCookie("at", accessT, timeExp, "/", "localhost", false, true)
	c.SetCookie("rt", rtB64, timeExp, "/", "localhost", true, true)
}
