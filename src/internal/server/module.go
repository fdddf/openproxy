package server

import "go.uber.org/fx"

// Module wires up the HTTP server lifecycle.
var Module = fx.Options(
	fx.Invoke(StartServer),
)
