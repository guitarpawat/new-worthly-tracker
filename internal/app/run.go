package app

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/guitarpawat/worthly-tracker/config"
	dbfiles "github.com/guitarpawat/worthly-tracker/db"
	frontendassets "github.com/guitarpawat/worthly-tracker/frontend"
	adapterdb "github.com/guitarpawat/worthly-tracker/internal/adapter/db"
	adapterlogger "github.com/guitarpawat/worthly-tracker/internal/adapter/logger"
	"github.com/guitarpawat/worthly-tracker/internal/adapter/repository"
	"github.com/guitarpawat/worthly-tracker/internal/service"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	linuxoptions "github.com/wailsapp/wails/v2/pkg/options/linux"
	macoptions "github.com/wailsapp/wails/v2/pkg/options/mac"
	windowsoptions "github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed appicon.png
var appIcon []byte

type runOptions struct {
	ConfigPath string
}

func Run(args []string) {
	runCfg, err := parseRunOptions(args)
	if err != nil {
		fail(err)
	}

	cfg, err := config.Load(runCfg.ConfigPath)
	if err != nil {
		fail(err)
	}

	logger, logFile, err := adapterlogger.New(cfg.Log)
	if err != nil {
		fail(err)
	}
	defer func() {
		_ = logFile.Close()
	}()
	logger.Info("starting application", "env", cfg.Env, "config_path", resolvedConfigPath(runCfg.ConfigPath), "db_path", resolvedDBPath(cfg.DB.Path), "log_path", resolvedLogPath(cfg.Log.Path), "log_level", cfg.Log.Level)

	database, err := adapterdb.Open(adapterdb.SQLiteConfig{Path: cfg.DB.Path})
	if err != nil {
		fail(err)
	}
	defer func() {
		_ = database.Close()
	}()

	if err := adapterdb.ApplyMigrations(context.Background(), database, dbfiles.FS); err != nil {
		fail(err)
	}

	recordRepository := repository.NewRecordSnapshotRepository(database)
	assetManagementRepository := repository.NewAssetManagementRepository(database)
	progressRepository := repository.NewProgressRepository(database)
	goalRepository := repository.NewGoalRepository(database)
	demoDataRepository := repository.NewDemoDataRepository(database, dbfiles.FS)
	recordService := service.NewRecordService(recordRepository)
	assetManagementService := service.NewAssetManagementService(assetManagementRepository)
	progressService := service.NewProgressService(progressRepository, goalRepository)
	goalService := service.NewGoalService(goalRepository)
	demoDataService := service.NewDemoDataService(demoDataRepository)
	wailsApp := New(logger, recordService, assetManagementService, progressService, goalService, demoDataService)

	if err := wails.Run(&options.App{
		Title:            cfg.Name,
		Width:            1440,
		Height:           960,
		MinWidth:         1100,
		MinHeight:        760,
		WindowStartState: options.Maximised,
		BackgroundColour: &options.RGBA{R: 250, G: 245, B: 235, A: 1},
		OnStartup:        wailsApp.Startup,
		Bind:             []any{wailsApp},
		AssetServer: &assetserver.Options{
			Assets: frontendassets.Files,
		},
		Windows: &windowsoptions.Options{
			DisableWindowIcon: false,
			Theme:             windowsoptions.SystemDefault,
			WindowClassName:   "WorthlyTracker",
		},
		Mac: &macoptions.Options{
			TitleBar: macoptions.TitleBarDefault(),
			About: &macoptions.AboutInfo{
				Title:   cfg.Name,
				Message: cfg.Name,
				Icon:    appIcon,
			},
		},
		Linux: &linuxoptions.Options{
			Icon:             appIcon,
			ProgramName:      "worthly-tracker",
			WebviewGpuPolicy: linuxoptions.WebviewGpuPolicyNever,
		},
	}); err != nil {
		fail(err)
	}
}

func parseRunOptions(args []string) (runOptions, error) {
	flags := flag.NewFlagSet("worthly-tracker", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var options runOptions
	flags.StringVar(&options.ConfigPath, "config", "", "path to config yaml file")
	if err := flags.Parse(args); err != nil {
		return runOptions{}, fmt.Errorf("parse args: %w", err)
	}

	return options, nil
}

func resolvedConfigPath(configPath string) string {
	if configPath == "" {
		return "./config/app.yaml"
	}
	return configPath
}

func resolvedDBPath(dbPath string) string {
	if dbPath == "" {
		return ":memory:"
	}
	return dbPath
}

func resolvedLogPath(logPath string) string {
	if logPath == "" {
		return "stdout only"
	}
	return logPath
}

func fail(err error) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	logger.Error("application startup failed", "err", err)
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
