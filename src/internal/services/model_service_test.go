package services

import (
	"path/filepath"
	"testing"

	"github.com/fdddf/openproxy/internal/dao"
	"github.com/fdddf/openproxy/internal/models"
	gormsqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// newTestDB opens a scratch SQLite database with the schema derived from the
// models, and points the generated DAO at it.
func newTestDB(t *testing.T, tables ...interface{}) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(
		gormsqlite.Open("file:"+filepath.Join(t.TempDir(), "test.db")),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)},
	)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(tables...); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	dao.SetDefault(db)
	return db
}

// A boolean that a caller deliberately sets to false must survive the round
// trip. gorm/gen implements Save as an upsert-Create, and GORM omits a field
// from the INSERT column list when it holds its zero value and carries a
// `default` tag disagreeing with that zero — which also leaves the column out
// of the ON CONFLICT update. `default:true` on Model.IsActive therefore made
// "false" unwritable through the API, on both create and update.
//
// Provider.HealthCheckEnabled is covered too. Its `default:false` agrees with
// the zero value so it was never affected, but the invariant is the same and
// worth pinning.
func TestFalseBooleansPersist(t *testing.T) {
	t.Run("model is_active on create", func(t *testing.T) {
		db := newTestDB(t, &models.Model{})

		m := models.Model{Name: "gpt-4o", ProviderID: 1, RealModel: "gpt-4o-2024-11-20", IsActive: false}
		if err := dao.Q.Model.Create(&m); err != nil {
			t.Fatalf("create model: %v", err)
		}

		if m.IsActive {
			t.Error("in-memory IsActive was rewritten to true by the database default")
		}
		if got := readBool(t, db, "select is_active from models where id = ?", m.ID); got {
			t.Error("IsActive=false was not persisted on create")
		}
	})

	t.Run("model is_active on update", func(t *testing.T) {
		db := newTestDB(t, &models.Model{})

		m := models.Model{Name: "gpt-4o", ProviderID: 1, RealModel: "gpt-4o-2024-11-20", IsActive: true}
		if err := dao.Q.Model.Create(&m); err != nil {
			t.Fatalf("create model: %v", err)
		}

		m.IsActive = false
		if err := dao.Q.Model.Save(&m); err != nil {
			t.Fatalf("save model: %v", err)
		}

		if got := readBool(t, db, "select is_active from models where id = ?", m.ID); got {
			t.Error("IsActive=false was not persisted on update")
		}
	})

	t.Run("provider health_check_enabled on update", func(t *testing.T) {
		db := newTestDB(t, &models.Provider{})

		p := models.Provider{Name: "openai", ApiKey: "sk-test", BaseURL: "https://api.openai.com/v1", HealthCheckEnabled: true}
		if err := dao.Q.Provider.Create(&p); err != nil {
			t.Fatalf("create provider: %v", err)
		}

		p.HealthCheckEnabled = false
		if err := dao.Q.Provider.Save(&p); err != nil {
			t.Fatalf("save provider: %v", err)
		}

		if got := readBool(t, db, "select health_check_enabled from providers where id = ?", p.ID); got {
			t.Error("HealthCheckEnabled=false was not persisted, health checks could not be turned off")
		}
	})
}

func readBool(t *testing.T, db *gorm.DB, query string, args ...interface{}) bool {
	t.Helper()

	var value bool
	if err := db.Raw(query, args...).Scan(&value).Error; err != nil {
		t.Fatalf("read column: %v", err)
	}
	return value
}
