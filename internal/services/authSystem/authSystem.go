package authsystem

import (
	"database/sql"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sater-151/AuthSystem/internal/apperror"
	"github.com/sater-151/AuthSystem/internal/database/postgresql"
	"github.com/sater-151/AuthSystem/internal/pkg/webhooks"
	"github.com/sater-151/AuthSystem/internal/utils"
	"github.com/sirupsen/logrus"
)

type AuthSystem interface {
	Login(guid string, host string, ip string) (aToken string, rToken string, err error)
	RefreshTokens(at string, rt string, userAgent string, ip string) (aToken string, rToken string, err error)
	CheckTokens(aToken string, rToken string) (err error)
	CompareRT(rToken, guid string) (ok bool, err error)
	Logout(guid string) (err error)
	GetGUID(aToken string) (guid string, err error)
}

type AuthSystemManager struct {
	db postgresql.Postgresql
	wh webhooks.WebHooks
}

func New(db postgresql.Postgresql, wh webhooks.WebHooks) *AuthSystemManager {
	authsystem := &AuthSystemManager{db: db, wh: wh}
	return authsystem
}

func (as *AuthSystemManager) Login(guid string, userAgent string, ip string) (aToken string, rToken string, err error) {
	logrus.Debug("starting authorization")
	aToken, rToken, err = utils.NewTokens(userAgent, guid)
	if err != nil {
		return aToken, rToken, err
	}
	err = as.db.LoginDB(guid, rToken, userAgent, ip)
	if err != nil {
		if err == sql.ErrNoRows {
			return aToken, rToken, apperror.ErrUnauthorized

		}
		return aToken, rToken, err
	}
	logrus.Debug("user logged")
	return aToken, rToken, nil
}

func (as *AuthSystemManager) RefreshTokens(at string, rt string, userAgent string, ip string) (aToken string, rToken string, err error) {
	logrus.Debug("refreshing tokens")

	guid, err := utils.GetGUIDFromJWT(at)
	if err != nil {
		return aToken, rToken, err
	}

	ok, err := as.CompareRT(rt, guid)
	if err != nil {
		return aToken, rToken, err
	}
	if !ok {
		return aToken, rToken, apperror.ErrUnauthorized
	}

	logrus.Debug("getting user info")
	oldUserAgent, oldUserIp, err := as.db.GetUserInfo(guid)
	if err != nil {
		return aToken, rToken, err
	}
	if userAgent != oldUserAgent {
		as.Logout(aToken)
		return aToken, rToken, apperror.ErrUnauthorized
	}
	if ip != oldUserIp {
		as.wh.SendMessageAboutAnotherIp()
	}

	logrus.Debug("generating new tokens")
	aToken, rToken, err = utils.NewTokens(userAgent, guid)
	if err != nil {
		return aToken, rToken, err
	}

	err = as.db.LoginDB(guid, rToken, userAgent, ip)
	if err != nil {
		return aToken, rToken, err
	}

	logrus.Debug("tokens refreshed")
	return aToken, rToken, nil
}

func (as *AuthSystemManager) CheckTokens(aToken string, rToken string) error {
	err := utils.CheckLinkTokens(aToken, rToken)
	if err != nil {
		if err.Error() == "token has invalid claims: token is expired" {
			return jwt.ErrTokenExpired
		}
		return err
	}
	return nil
}

func (as *AuthSystemManager) CompareRT(rToken, guid string) (bool, error) {
	rTokenBcryptNew, err := as.db.GetBcrypt(rToken)
	if err != nil {
		return false, err
	}
	rTokenOld, err := as.db.GetToken(guid)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, apperror.ErrUnauthorized
		}
		return false, err
	}
	if rTokenBcryptNew != rTokenOld {
		return false, nil
	}
	return true, nil
}

func (as *AuthSystemManager) Logout(aToken string) error {
	guid, err := utils.GetGUIDFromJWT(aToken)
	if err != nil {
		return err
	}
	return as.db.DeleteUser(guid)
}

func (as *AuthSystemManager) GetGUID(aToken string) (string, error) {
	return utils.GetGUIDFromJWT(aToken)
}
