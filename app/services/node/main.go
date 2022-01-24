package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ardanlabs/blockchain/app/services/node/handlers"
	"github.com/ardanlabs/blockchain/foundation/blockchain"
	"github.com/ardanlabs/blockchain/foundation/logger"
	"github.com/ardanlabs/conf/v3"
	"go.uber.org/zap"
)

// build is the git version of this program. It is set using build flags in the makefile.
var build = "develop"

func main() {

	// Construct the application logger.
	log, err := logger.New("BAR-LED")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer log.Sync()

	// Perform the startup and shutdown sequence.
	if err := run(log); err != nil {
		log.Errorw("startup", "ERROR", err)
		log.Sync()
		os.Exit(1)
	}
}

func run(log *zap.SugaredLogger) error {

	// =========================================================================
	// Configuration

	cfg := struct {
		conf.Version
		Web struct {
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:10s"`
			IdleTimeout     time.Duration `conf:"default:120s"`
			ShutdownTimeout time.Duration `conf:"default:20s"`
			DebugHost       string        `conf:"default:0.0.0.0:7080"`
			PublicHost      string        `conf:"default:0.0.0.0:8080"`
			PrivateHost     string        `conf:"default:0.0.0.0:9080"`
		}
		Node struct {
			MinerAccount  string   `conf:"default:miner1"`
			DBPath        string   `conf:"default:zblock/blocks.db"`
			KnownPeers    []string `conf:"default:0.0.0.0:9080;0.0.0.0:9180"`
			Reward        uint     `conf:"default:700"`
			Difficulty    int      `conf:"default:6"`
			TransPerBlock int      `conf:"default:2"`
		}
	}{
		Version: conf.Version{
			Build: build,
			Desc:  "copyright information here",
		},
	}

	const prefix = "BARLED"
	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}
		return fmt.Errorf("parsing config: %w", err)
	}

	// =========================================================================
	// App Starting

	log.Infow("starting service", "version", build)
	defer log.Infow("shutdown complete")

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}
	log.Infow("startup", "config", out)

	// =========================================================================
	// Blockchain Support

	peerSet := blockchain.NewPeerSet()
	for _, host := range cfg.Node.KnownPeers {
		peerSet.Add(blockchain.NewPeer(host))
	}

	ev := func(v string, args ...interface{}) {
		s := fmt.Sprintf(v, args...)
		log.Infow(s, "traceid", "11111111-1111-1111-1111-111111111111")
	}

	bc, err := blockchain.New(blockchain.Config{
		MinerAccount:  cfg.Node.MinerAccount,
		Host:          cfg.Web.PrivateHost,
		DBPath:        cfg.Node.DBPath,
		KnownPeers:    peerSet,
		Reward:        cfg.Node.Reward,
		Difficulty:    cfg.Node.Difficulty,
		TransPerBlock: cfg.Node.TransPerBlock,
		EvHandler:     ev,
	})
	if err != nil {
		return err
	}
	defer bc.Shutdown()

	// =========================================================================
	// Start Debug Service

	log.Infow("startup", "status", "debug v1 router started", "host", cfg.Web.DebugHost)

	// The Debug function returns a mux to listen and serve on for all the debug
	// related endpoints. This includes the standard library endpoints.

	// Construct the mux for the debug calls.
	debugMux := handlers.DebugMux(build, log)

	// Start the service listening for debug requests.
	// Not concerned with shutting this down with load shedding.
	go func() {
		if err := http.ListenAndServe(cfg.Web.DebugHost, debugMux); err != nil {
			log.Errorw("shutdown", "status", "debug v1 router closed", "host", cfg.Web.DebugHost, "ERROR", err)
		}
	}()

	// =========================================================================
	// Service Start/Stop Support

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// =========================================================================
	// Start Public Service

	log.Infow("startup", "status", "initializing V1 public API support")

	// Construct the mux for the public API calls.
	publicMux := handlers.PublicMux(handlers.MuxConfig{
		Shutdown: shutdown,
		Log:      log,
		BC:       bc,
	})

	// Construct a server to service the requests against the mux.
	public := http.Server{
		Addr:         cfg.Web.PublicHost,
		Handler:      publicMux,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
		ErrorLog:     zap.NewStdLog(log.Desugar()),
	}

	// Start the service listening for api requests.
	go func() {
		log.Infow("startup", "status", "public api router started", "host", public.Addr)
		serverErrors <- public.ListenAndServe()
	}()

	// =========================================================================
	// Start Private Service

	log.Infow("startup", "status", "initializing V1 private API support")

	// Construct the mux for the private API calls.
	privateMux := handlers.PrivateMux(handlers.MuxConfig{
		Shutdown: shutdown,
		Log:      log,
		BC:       bc,
	})

	// Construct a server to service the requests against the mux.
	private := http.Server{
		Addr:         cfg.Web.PrivateHost,
		Handler:      privateMux,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
		ErrorLog:     zap.NewStdLog(log.Desugar()),
	}

	// Start the service listening for api requests.
	go func() {
		log.Infow("startup", "status", "private api router started", "host", private.Addr)
		serverErrors <- private.ListenAndServe()
	}()

	// =========================================================================
	// Shutdown

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Infow("shutdown", "status", "shutdown started", "signal", sig)
		defer log.Infow("shutdown", "status", "shutdown complete", "signal", sig)

		// Give outstanding requests a deadline for completion.
		ctx, cancelPub := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancelPub()

		// Asking listener to shut down and shed load.
		log.Infow("shutdown", "status", "shutdown private API started")
		if err := private.Shutdown(ctx); err != nil {
			private.Close()
			return fmt.Errorf("could not stop private service gracefully: %w", err)
		}

		// Give outstanding requests a deadline for completion.
		ctx, cancelPri := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancelPri()

		// Asking listener to shut down and shed load.
		log.Infow("shutdown", "status", "shutdown public API started")
		if err := public.Shutdown(ctx); err != nil {
			public.Close()
			return fmt.Errorf("could not stop public service gracefully: %w", err)
		}
	}

	return nil
}
