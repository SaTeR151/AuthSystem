package rest

import (
	"encoding/base64"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sater-151/AuthSystem/internal/apperror"
	"github.com/sater-151/AuthSystem/internal/controller/rest/dto"
	"github.com/sater-151/AuthSystem/internal/controller/rest/restutils"
	authsystem "github.com/sater-151/AuthSystem/internal/services/authSystem"
	"github.com/sirupsen/logrus"
)

// Authorization godoc
//
// @Summary		User authorization
// @Description	Генерация access и refresh токенов для пользователя с указанным guid
// @Tags		Auth
// @Param       guid   query      string  true  "user guid/id" default(090bb747-d6d3-4067-a1da-2b83726eb24d)
// @Success		201
// @Header      201 {Cookie}  at  "access token. Время жизни токена 60 секунд. Время жизни Cookie 30 дней."
// @Header      201 {Cookie}  rt  "refresh token. Время жизни Cookie 30 дней."
// @Failure		400	{object}	map[string]string
// @Failure		401	{object}	map[string]string
// @Failure		500	{object}	map[string]string
// @Router		/api/login [post]
func Login(as authsystem.AuthSystem) gin.HandlerFunc {
	return func(c *gin.Context) {
		logrus.Info("getting tokens")
		guid := c.Query("guid")
		if guid == "" {
			logrus.Error(apperror.ErrGUIDRequired)
			restutils.Error(c, apperror.ErrGUIDRequired.Error(), http.StatusBadRequest)
			return
		}

		aToken, rToken, err := as.Login(guid, c.Request.Header.Get("User-Agent"), c.ClientIP())
		if err != nil {
			if err == apperror.ErrUnauthorized {
				logrus.Warn(err)
				restutils.Error(c, err.Error(), http.StatusUnauthorized)
				return
			}
			logrus.Error(err)
			restutils.Error(c, err.Error(), http.StatusInternalServerError)
			return
		}

		restutils.SetCookieTokens(c, aToken, rToken)
		c.Status(http.StatusCreated)
		logrus.Info("tokens have been sent")
	}
}

// RefreshTokens godoc
//
// @Summary		Refresh tokens
// @Description	Генерация новых access и refresh токенов на основе guid в access токене
// @Tags		Auth
// @Success		201
// @Header      201 {Cookie}  at  "access token. Время жизни токена 60 секунд. Время жизни Cookie 30 дней."
// @Header      201 {Cookie}  rt  "refresh token. Время жизни Cookie 30 дней."
// @Failure		400	{object}	map[string]string
// @Failure		401	{object}	map[string]string
// @Failure		500	{object}	map[string]string
// @Router		/api/refresh [post]
func Refresh(as authsystem.AuthSystem) gin.HandlerFunc {
	return func(c *gin.Context) {
		logrus.Info("starting refreshing tokens")
		rtCookie, err := c.Request.Cookie("rt")
		if err != nil {
			logrus.Error(err)
			restutils.Error(c, apperror.ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}
		rTokenCookie, err := url.QueryUnescape(rtCookie.Value)
		if err != nil {
			logrus.Error(err)
			restutils.Error(c, apperror.ErrIncorrectRefreshToken.Error(), http.StatusBadRequest)
			return
		}
		gettingRTBase64, err := base64.StdEncoding.DecodeString(rTokenCookie)
		if err != nil {
			logrus.Error(err)
			restutils.Error(c, apperror.ErrIncorrectRefreshToken.Error(), http.StatusBadRequest)
			return
		}

		atCookie, err := c.Request.Cookie("at")
		if err != nil {
			logrus.Error(err)
			restutils.Error(c, apperror.ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}
		if err = as.CheckTokens(atCookie.Value, string(gettingRTBase64)); err != nil {
			if err != jwt.ErrTokenExpired {
				logrus.Error(err)
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
		}
		aToken, rToken, err := as.RefreshTokens(atCookie.Value, string(gettingRTBase64), c.Request.Header.Get("User-Agent"), c.ClientIP())
		if err != nil {
			if err == apperror.ErrUnauthorized {
				logrus.Warn(err)
				restutils.Error(c, apperror.ErrUnauthorized.Error(), http.StatusUnauthorized)
				return
			}
			logrus.Error(err)
			restutils.Error(c, "", http.StatusInternalServerError)
			return
		}
		restutils.SetCookieTokens(c, aToken, rToken)
		c.Status(http.StatusCreated)

		logrus.Info("tokens refreshed")
	}
}

// Deauthorization godoc
//
// @Summary		User deauthorization
// @Security 	AccessToken
// @Security 	RefreshToken
// @Description	Деавторизация пользователя на основе guid из access токена. ВНИМАНИЕ! Guid пользователя будет удалено из БД
// @Tags		Auth
// @Success		204
// @Failure		401	{object}	map[string]string
// @Failure		500	{object}	map[string]string
// @Router		/api/auth/logout [post]
func Deauthorization(as authsystem.AuthSystem) gin.HandlerFunc {
	return func(c *gin.Context) {
		logrus.Info("starting logout")
		atCookie, err := c.Request.Cookie("at")
		if err != nil {
			logrus.Error(err)
			restutils.Error(c, apperror.ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}

		err = as.Logout(atCookie.Value)
		if err != nil {
			logrus.Warn(err)
			if err != apperror.ErrUnauthorized {
				restutils.Error(c, "", http.StatusInternalServerError)
				return
			}
		}
		c.Status(http.StatusNoContent)
		logrus.Info("logout finished")
	}
}

// GetGUID godoc
//
// @Summary		Get user's guid
// @Security 	AccessToken
// @Security 	RefreshToken
// @Description	получение guid пользователя из полученного access токена
// @Tags		Get
// @Success		200	{object} 	dto.GUID
// @Failure		401	{object}	map[string]string
// @Failure		500	{object}	map[string]string
// @Router		/api/auth/guid [get]
func GetGUID(as authsystem.AuthSystem) gin.HandlerFunc {
	return func(c *gin.Context) {
		logrus.Info("getting guid")
		atCookie, err := c.Request.Cookie("at")
		if err != nil {
			logrus.Error(err)
			restutils.Error(c, apperror.ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}

		var guid dto.GUID
		guid.Guid, err = as.GetGUID(atCookie.Value)
		if err != nil {
			logrus.Error(err)
			restutils.Error(c, "", http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, guid)
	}
}
