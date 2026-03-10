package controllers

import (
	"go.uber.org/fx"
)

// Module provides the controller for the application
var ControllerModule = fx.Options(
	fx.Provide(
		NewController,
	),
)
