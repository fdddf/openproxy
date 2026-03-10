package cmd

import (
	"fmt"
	"os"

	"github.com/fdddf/openproxy/internal/services"
	"github.com/spf13/cobra"
)

func newMigrateCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigrations(configPath, force)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force re-run migrations (sets FORCE_MIGRATE=1)")
	return cmd
}

func runMigrations(cfgPath string, force bool) error {
	if force {
		// signal to underlying services/migrations that a forced run is requested
		if err := os.Setenv("FORCE_MIGRATE", "1"); err != nil {
			return fmt.Errorf("set force env: %w", err)
		}
	}

	configService := services.NewConfigService()
	if err := configService.LoadConfig(cfgPath); err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	dbService := services.NewDatabaseService(configService)
	if err := dbService.InitDatabase(); err != nil {
		return fmt.Errorf("init database: %w", err)
	}

	if gormDB := dbService.GetDB(); gormDB != nil {
		if sqlDB, err := gormDB.DB(); err == nil {
			defer sqlDB.Close()
		}
	}

	return nil
}
