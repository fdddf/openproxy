package services

import (
	"database/sql"
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/fdddf/openproxy/common"
	"github.com/fdddf/openproxy/internal/dao"
	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DatabaseServiceImpl implements the DatabaseService interface
type DatabaseServiceImpl struct {
	db            *gorm.DB
	dao           *dao.Query
	configService ConfigService
}

// NewDatabaseService creates a new DatabaseService instance
func NewDatabaseService(configService ConfigService) DatabaseService {
	return &DatabaseServiceImpl{
		db:            nil,
		configService: configService,
	}
}

// GetDB returns the database instance
func (d *DatabaseServiceImpl) GetDB() *gorm.DB {
	return d.db
}

// GetDAO returns the generated gorm DAO query helpers
func (d *DatabaseServiceImpl) GetDAO() *dao.Query {
	return d.dao
}

// InitDatabase initializes the database connection
func (d *DatabaseServiceImpl) InitDatabase() error {
	cfg := d.configService.GetConfig()
	dsn, err := buildPostgresDSN(cfg.Database)
	if err != nil {
		return err
	}

	gormDB, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return fmt.Errorf("init sql db handle: %w", err)
	}

	if err := runMigrations(sqlDB, cfg.Database); err != nil {
		return err
	}

	d.db = gormDB
	dao.SetDefault(gormDB)
	d.dao = dao.Q
	return nil
}

func buildPostgresDSN(dbConfig common.DatabaseConfig) (string, error) {
	if dbConfig.Host == "" || dbConfig.Name == "" || dbConfig.User == "" {
		return "", fmt.Errorf("database configuration missing host, user, or name")
	}

	port := dbConfig.Port
	if port == 0 {
		port = 5432
	}

	sslMode := dbConfig.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	u := url.URL{
		Scheme: "postgres",
		Host:   fmt.Sprintf("%s:%d", dbConfig.Host, port),
		Path:   dbConfig.Name,
	}
	if dbConfig.User != "" {
		if dbConfig.Password != "" {
			u.User = url.UserPassword(dbConfig.User, dbConfig.Password)
		} else {
			u.User = url.User(dbConfig.User)
		}
	}

	// ensure we always specify sslmode
	query := url.Values{}
	query.Set("sslmode", sslMode)
	u.RawQuery = query.Encode()

	return u.String(), nil
}

func runMigrations(sqlDB *sql.DB, dbConfig common.DatabaseConfig) error {
	migrationsPath := dbConfig.MigrationsPath
	if migrationsPath == "" {
		migrationsPath = "./migrations"
	}

	sourceURL, err := resolveMigrationsSource(migrationsPath)
	if err != nil {
		return err
	}

	driver, err := migratepg.WithInstance(sqlDB, &migratepg.Config{
		DatabaseName: dbConfig.Name,
	})
	if err != nil {
		return fmt.Errorf("create postgres migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(sourceURL, dbConfig.Name, driver)
	if err != nil {
		return fmt.Errorf("init migrations: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}

func resolveMigrationsSource(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve migrations path: %w", err)
	}
	return "file://" + filepath.ToSlash(absPath), nil
}
