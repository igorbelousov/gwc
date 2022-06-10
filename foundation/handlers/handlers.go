package handlers

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/igorbelousov/gwc/foundation/web"
	"go.uber.org/zap"
)

type APIMuxConfig struct {
	Shutdown chan os.Signal
	Log      *zap.SugaredLogger
	// Metrics *metrics.Metrics
	// Auth *auth.Auth
}

func APIMux(cfg APIMuxConfig) *web.App {
	app := web.NewApp(cfg.Shutdown)

	h := func(w http.ResponseWriter, r *http.Request) {
		status := struct {
			Status string
		}{
			Status: "OK",
		}
		json.NewEncoder(w).Encode(status)
	}

	app.Handle(http.MethodGet, "/test", h)

	return app
}
