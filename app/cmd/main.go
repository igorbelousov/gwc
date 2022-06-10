package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/ardanlabs/conf"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var build = "develop"

func main() {

	log, err := initLog("GWC-service")
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	if err := run(log); err != nil {
		log.Errorw("Start", "Error:", err)
		os.Exit(1)
	}
}

func run(log *zap.SugaredLogger) error {

	// // ============================================================
	// // GOMAXPROCS

	if _, err := maxprocs.Set(); err != nil {
		fmt.Println("maxprocs: %w", err)
		// os.Exit(1)
	}
	log.Infow("start", "GOMAXPROCX", runtime.GOMAXPROCS(0))

	// ============================================================
	// CONFIGURATION

	cfg := struct {
		conf.Version
		Web struct {
			APIHost         string        `conf:"default:0.0.0.0:3000"`
			DebugHost       string        `conf:"default:0.0.0.0:4000"`
			ReadTimeout     time.Duration `conf:"default:5s"`
			writeTimeout    time.Duration `conf:"default:10s"`
			idleTimeout     time.Duration `conf:"default:120s"`
			ShutdownTimeout time.Duration `conf:"default:20s"`
		}
	}{
		Version: conf.Version{
			SVN:  build,
			Desc: "Copyright info",
		},
	}

	const prefix = "GWC"
	help, err := conf.ParseOSArgs(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}
		return fmt.Errorf("parsing conf error: %w", err)
	}

	// ============================================================
	// APP STARTER
	log.Infow("start service", "version", build)
	defer log.Infow("shutdown complete")

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generation config output,  %w", err)

	}
	log.Infow("startup", "config", out)

	// ============================================================

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      nil,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.writeTimeout,
		IdleTimeout:  cfg.Web.idleTimeout,
		ErrorLog:     zap.NewStdLog(log.Desugar()),
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Infow("startup", "status", "api router started", "host", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error %w", err)

	case sig := <-shutdown:
		log.Infow("shutdown", "status", "shutdown started", "signal", sig)
		defer log.Infow("shutdown", "status", "shutdown complete", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		if err := api.Shutdown(ctx); err != nil {
			api.Close()
			return fmt.Errorf("could not stop server gracefully:  %w", err)
		}
	}

	return nil
}

func initLog(service string) (*zap.SugaredLogger, error) {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.DisableStacktrace = true
	config.InitialFields = map[string]interface{}{
		"service": service,
	}

	log, err := config.Build()
	if err != nil {
		fmt.Println("Error construct logger:", err)
		os.Exit(1)
	}

	return log.Sugar(), nil

}
