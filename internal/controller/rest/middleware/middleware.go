package middleware

import (
	"encoding/base64"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sater-151/AuthSystem/internal/apperror"
	"github.com/sater-151/AuthSystem/internal/controller/rest/restutils"
	authsystem "github.com/sater-151/AuthSystem/internal/services/authSystem"
	"github.com/sirupsen/logrus"
)

func CheckAuthorization(as authsystem.AuthSystem) gin.HandlerFunc {
	return func(c *gin.Context) {
		logrus.Info("checking authorization")
		rtCookie, err := c.Request.Cookie("rt")
		if err != nil {
			logrus.Error(err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		rTokenCookie, err := url.QueryUnescape(rtCookie.Value)
		if err != nil {
			logrus.Error(err)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		gettingRTBase64, err := base64.StdEncoding.DecodeString(rTokenCookie)
		if err != nil {
			logrus.Error(err)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		accessTCokie, err := c.Request.Cookie("at")
		if err != nil {
			logrus.Error(err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if err = as.CheckTokens(accessTCokie.Value, string(gettingRTBase64)); err != nil {
			if err != jwt.ErrTokenExpired {
				logrus.Error(err)
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			logrus.Info("access token expired")
			aToken, rToken, err := as.RefreshTokens(accessTCokie.Value, string(gettingRTBase64), c.Request.Header.Get("User-Agent"), c.ClientIP())
			if err != nil {
				logrus.Warn(err)
				if err == apperror.ErrUnauthorized {
					c.AbortWithStatus(http.StatusUnauthorized)
					return
				}
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			restutils.SetCookieTokens(c, aToken, rToken)
			c.Next()
			return
		}
		c.Next()
		return
	}
}
