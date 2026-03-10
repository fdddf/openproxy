package cmd

import (
	"context"
	"time"

	"github.com/fdddf/openproxy/internal/controllers"
	"github.com/fdddf/openproxy/internal/server"
	"github.com/fdddf/openproxy/internal/services"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

func newServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the gptproxy HTTP server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer(cmd.Context(), configPath)
		},
	}
}

func runServer(ctx context.Context, cfgPath string) error {
	app := fx.New(
		fx.Supply(
			fx.Annotate(cfgPath, fx.ResultTags(`name:"configPath"`)),
		),
		services.Module,
		controllers.ControllerModule,
		server.Module,
	)

	if err := app.Start(ctx); err != nil {
		return err
	}

	<-app.Done()

	stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return app.Stop(stopCtx)
}
