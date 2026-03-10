package cmd

import (
	"log"

	"github.com/fdddf/openproxy/internal/models"
	"github.com/spf13/cobra"
	"gorm.io/gen"
)

func newGenDAOCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "gendao",
		Short: "Generate gorm/gen DAO layer under internal/dao",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runGenDAO()
		},
	}
}

func runGenDAO() error {
	g := gen.NewGenerator(gen.Config{
		OutPath: "internal/dao",
		Mode:    gen.WithoutContext | gen.WithDefaultQuery,
	})

	g.ApplyBasic(
		models.User{},
		models.APIKey{},
		models.Log{},
		models.Model{},
		models.Provider{},
		models.Request{},
		models.Setting{},
		models.OauthSession{},
	)

	g.Execute()
	log.Println("DAO generated under internal/dao")
	return nil
}
