package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/alecthomas/kong"

	"github.com/dedene/memelink-cli/internal/api"
	"github.com/dedene/memelink-cli/internal/config"
	"github.com/dedene/memelink-cli/internal/outfmt"
	"github.com/dedene/memelink-cli/internal/ui"
)

// RootFlags are global flags available to all commands.
type RootFlags struct {
	Color   string `help:"Color output: auto|always|never" default:"auto" enum:"auto,always,never"`
	JSON    bool   `help:"JSON output" default:"false"`
	Verbose bool   `help:"Verbose logging" default:"false"`
	NoInput bool   `help:"Never prompt; fail instead" name:"no-input" default:"false"`
	Force   bool   `help:"Skip confirmations" default:"false"`
}

// CLI is the top-level Kong command struct.
type CLI struct {
	RootFlags `embed:""`

	Version    kong.VersionFlag `help:"Print version and exit"`
	VersionCmd VersionCmd       `cmd:"" name:"version" help:"Print version info"`
	Generate   GenerateCmd      `cmd:"" name:"generate" aliases:"gen,g" default:"withargs" help:"Generate a meme"`
	Templates  TemplatesCmd     `cmd:"" name:"templates" aliases:"ls" help:"List or view templates"`
	Fonts      FontsCmd         `cmd:"" name:"fonts" help:"List or view fonts"`
	Config     ConfigCmd        `cmd:"" name:"config" help:"Manage configuration"`
}

// Execute parses CLI args, sets up context, and runs the matched command.
func Execute(args []string) (err error) {
	cli := &CLI{}
	parser, err := kong.New(
		cli,
		kong.Name("memelink"),
		kong.Description("Generate memes from the terminal"),
		kong.ConfigureHelp(helpOptions()),
		kong.Help(helpPrinter),
		kong.Vars{"version": VersionString()},
		kong.Writers(os.Stdout, os.Stderr),
		kong.Exit(func(code int) { panic(exitPanic{code: code}) }),
	)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			if ep, ok := r.(exitPanic); ok {
				if ep.code == 0 {
					err = nil
					return
				}
				err = &ExitError{Code: ep.code, Err: errors.New("exited")}
				return
			}
			panic(r)
		}
	}()

	kctx, err := parser.Parse(args)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return &ExitError{Code: 2, Err: err}
	}

	// Verbose logging
	logLevel := slog.LevelWarn
	if cli.Verbose {
		logLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	})))

	// Output mode
	mode := outfmt.Mode{JSON: cli.JSON}
	ctx := context.Background()
	ctx = outfmt.WithMode(ctx, mode)

	// UI printer -- force no color in JSON mode
	uiColor := cli.Color
	if outfmt.IsJSON(ctx) {
		uiColor = "never"
	}
	u, uiErr := ui.New(ui.Options{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Color:  uiColor,
	})
	if uiErr != nil {
		return uiErr
	}
	ctx = ui.WithUI(ctx, u)

	// Config
	cfgPath, _ := config.ConfigPath()
	cfg, cfgErr := config.Load(cfgPath)
	if cfgErr != nil {
		slog.Warn("loading config", "error", cfgErr)
		cfg = &config.Config{}
	}
	ctx = config.WithConfig(ctx, cfg)

	// API client
	client := api.NewClient(api.ClientOptions{
		APIKey:    os.Getenv("MEMEGEN_API_KEY"),
		Verbose:   cli.Verbose,
		UserAgent: "memelink-cli/" + version,
	})
	ctx = api.WithClient(ctx, client)

	// Bind context + root flags to Kong
	kctx.BindTo(ctx, (*context.Context)(nil))
	kctx.Bind(&cli.RootFlags)

	return kctx.Run()
}
