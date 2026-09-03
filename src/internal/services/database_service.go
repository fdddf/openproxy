package services

import (
	"database/sql"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/fdddf/openproxy/common"
	"github.com/fdddf/openproxy/internal/dao"
	"github.com/fdddf/openproxy/migrations"
	gormsqlite "github.com/glebarez/sqlite"
	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	gormpostgres "gorm.io/driver/postgres"
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

// InitDatabase opens the configured database, applies migrations, and wires up
// the generated DAO.
func (d *DatabaseServiceImpl) InitDatabase() error {
	cfg := d.configService.GetConfig()
	dbConfig := cfg.Database

	var (
		gormDB *gorm.DB
		err    error
	)

	gormCfg := &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}

	switch dbConfig.Driver {
	case common.DriverSQLite:
		gormDB, err = openSQLite(dbConfig, gormCfg)
	case common.DriverPostgres:
		gormDB, err = openPostgres(dbConfig, gormCfg)
	default:
		return fmt.Errorf("unsupported database driver %q (want %q or %q)",
			dbConfig.Driver, common.DriverSQLite, common.DriverPostgres)
	}
	if err != nil {
		return err
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return fmt.Errorf("init sql db handle: %w", err)
	}

	if dbConfig.Driver == common.DriverSQLite {
		// SQLite serialises writes at the file level. WAL plus a busy timeout
		// (set in the DSN) lets readers run concurrently, but a single writer
		// connection avoids SQLITE_BUSY under load.
		sqlDB.SetMaxOpenConns(1)
		sqlDB.SetMaxIdleConns(1)
	}

	if err := runMigrations(sqlDB, dbConfig); err != nil {
		return err
	}

	d.db = gormDB
	dao.SetDefault(gormDB)
	d.dao = dao.Q
	return nil
}

func openSQLite(dbConfig common.DatabaseConfig, gormCfg *gorm.Config) (*gorm.DB, error) {
	path := dbConfig.Path
	if path == "" {
		path = common.DefaultSQLitePath
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve sqlite path %q: %w", path, err)
	}
	if dir := filepath.Dir(absPath); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create sqlite directory %q: %w", dir, err)
		}
	}

	// WAL keeps readers from blocking the writer; busy_timeout makes the ones
	// that do collide wait instead of failing outright.
	dsn := "file:" + absPath + "?" + strings.Join([]string{
		"_pragma=journal_mode(WAL)",
		"_pragma=busy_timeout(5000)",
		"_pragma=synchronous(NORMAL)",
		"_pragma=foreign_keys(ON)",
	}, "&")

	gormDB, err := gorm.Open(gormsqlite.Open(dsn), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("open sqlite %s: %w", absPath, err)
	}
	return gormDB, nil
}

func openPostgres(dbConfig common.DatabaseConfig, gormCfg *gorm.Config) (*gorm.DB, error) {
	dsn, err := buildPostgresDSN(dbConfig)
	if err != nil {
		return nil, err
	}

	gormDB, err := gorm.Open(gormpostgres.Open(dsn), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	return gormDB, nil
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
	source, err := migrationSource(dbConfig)
	if err != nil {
		return err
	}

	if dbConfig.Driver == common.DriverSQLite {
		return runSQLiteMigrations(sqlDB, source)
	}

	src, err := iofs.New(source, ".")
	if err != nil {
		return fmt.Errorf("open migrations: %w", err)
	}
	defer src.Close()

	driver, err := migratepg.WithInstance(sqlDB, &migratepg.Config{
		DatabaseName: dbConfig.Name,
	})
	if err != nil {
		return fmt.Errorf("create postgres migration driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, dbConfig.Name, driver)
	if err != nil {
		return fmt.Errorf("init migrations: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}

// migrationSource returns the migration set for the configured driver. It uses
// the migrations embedded in the binary unless an explicit path is configured.
func migrationSource(dbConfig common.DatabaseConfig) (fs.FS, error) {
	if path := dbConfig.MigrationsPath; path != "" {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("resolve migrations path: %w", err)
		}
		return os.DirFS(absPath), nil
	}

	sub, err := fs.Sub(migrations.FS, dbConfig.Driver)
	if err != nil {
		return nil, fmt.Errorf("locate embedded migrations for %q: %w", dbConfig.Driver, err)
	}
	return sub, nil
}
