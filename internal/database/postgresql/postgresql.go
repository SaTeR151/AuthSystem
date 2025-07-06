package postgresql

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sater-151/AuthSystem/internal/apperror"
	"github.com/sater-151/AuthSystem/internal/config"
	"github.com/sirupsen/logrus"
)

type Postgresql interface {
	MigrationUp() (err error)
	MigrationDown() (err error)
	LoginDB(guid string, rt string, userAgent string, ip string) (err error)
	UpdateRT(guid string, rt string) (err error)
	GetBcrypt(rToken string) (rTokenBcrypt string, err error)
	GetToken(guid string) (rToken string, err error)
	DeleteUser(guid string) (err error)
	GetUserInfo(guid string) (userAgent string, userIp string, err error)
}

type PostgresqlManager struct {
	db   *sql.DB
	hash string
}

func Open(config config.PostgresqlConfig) (*PostgresqlManager, func() error, error) {
	var err error
	connInfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		config.Host,
		config.User,
		config.Pass,
		config.Dbname,
		config.Port,
		config.Sslmode)

	db, err := sql.Open("pgx", connInfo)
	if err != nil {
		return nil, nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, nil, err
	}
	DB := &PostgresqlManager{db: db}
	var ok bool
	DB.hash, ok = os.LookupEnv("BCRYPYHASH")
	if !ok {
		logrus.Warn("bcrypt hash isn't found")
	}
	return DB, db.Close, nil
}

func (db *PostgresqlManager) MigrationUp() error {
	driver, err := postgres.WithInstance(db.db, &postgres.Config{})
	if err != nil {
		return err
	}
	migrator, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		return err
	}
	err = migrator.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func (db *PostgresqlManager) MigrationDown() error {
	driver, err := postgres.WithInstance(db.db, &postgres.Config{})
	if err != nil {
		return err
	}
	migrator, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		return err
	}
	err = migrator.Down()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func (db *PostgresqlManager) LoginDB(guid string, rt string, userAgent string, ip string) error {
	logrus.Debug("set refresh token")
	res, err := db.db.Exec("UPDATE users_auth SET refresh_t=crypt($1, $2), user_agent=$3, user_ip=$4 WHERE user_id=$5 RETURNING refresh_t", rt, db.hash, userAgent, ip, guid)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	logrus.Debug("refresh token has beeb set")
	return nil
}

func (db *PostgresqlManager) UpdateRT(guid string, rt string) error {
	logrus.Debug("set refresh token")
	res, err := db.db.Exec("UPDATE users_auth SET refresh_t=crypt($1, $2) WHERE user_id=$3 RETURNING refresh_t", rt, db.hash, guid)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	logrus.Debug("refresh token has beeb set")
	return nil
}

func (db *PostgresqlManager) GetBcrypt(rToken string) (string, error) {
	var rTokenBcrypt sql.NullString
	err := db.db.QueryRow("SELECT crypt($1, $2)", rToken, db.hash).Scan(&rTokenBcrypt)
	if err != nil {
		return "", err
	}
	return rTokenBcrypt.String, nil
}

func (db *PostgresqlManager) GetToken(guid string) (string, error) {
	var rtDB sql.NullString
	err := db.db.QueryRow("SELECT refresh_t FROM users_auth WHERE user_id=$1", guid).Scan(&rtDB)
	if err != nil {
		return "", err
	}
	return rtDB.String, nil
}

func (db *PostgresqlManager) DeleteUser(guid string) error {
	res, err := db.db.Exec("DELETE FROM users_auth WHERE user_id=$1", guid)
	if err != nil {
		return err
	}
	row, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if row == 0 {
		return apperror.ErrUnauthorized
	}
	logrus.Info("user deauthorized")
	return nil
}

func (db *PostgresqlManager) GetUserInfo(guid string) (string, string, error) {
	var userAgent, userIp string
	err := db.db.QueryRow("SELECT user_agent, user_ip FROM users_auth WHERE user_id=$1", guid).Scan(&userAgent, &userIp)
	if err != nil {
		return "", "", err
	}
	return userAgent, userIp, nil
}
