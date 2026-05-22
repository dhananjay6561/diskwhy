package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/dhananjay6561/diskwhy/internal/config"
	"github.com/dhananjay6561/diskwhy/internal/errtype"
	"github.com/dhananjay6561/diskwhy/internal/exitcode"
	"github.com/spf13/cobra"
)

// GlobalConfig is populated by PersistentPreRunE before any subcommand runs.
var GlobalConfig *config.Config

// Execute is the single entry point called from main. It wires the
// cancellation context, installs signal handlers for SIGINT and SIGTERM,
// runs the cobra command tree, and exits with the appropriate code.
//
// Signal handler contract: the goroutine calls cancel() and returns
// immediately. It does not block, join, or wait on any other goroutine.
// This keeps the handler fast and deadlock-free.
func Execute(ver, buildCommit string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	// interrupted is buffered so the goroutine never blocks after cancel().
	interrupted := make(chan struct{}, 1)
	go func() {
		select {
		case <-sigCh:
			interrupted <- struct{}{}
			cancel()
		case <-ctx.Done():
		}
	}()

	root := buildRootCmd(ver, buildCommit)
	err := root.ExecuteContext(ctx)

	// Signal check takes priority over any RunE error.
	select {
	case <-interrupted:
		fmt.Fprintln(os.Stderr, "Cancelled")
		os.Exit(exitcode.Interrupted)
	default:
	}

	if err != nil {
		var coded *errtype.CodedError
		if errors.As(err, &coded) {
			fmt.Fprintln(os.Stderr, coded.Error())
			os.Exit(coded.Code)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(cobraExitCode(err))
	}
}

// cobraExitCode maps cobra/pflag error messages to the appropriate exit code.
// Cobra does not expose typed flag errors, so message-prefix matching is the
// only available mechanism.
func cobraExitCode(err error) int {
	msg := err.Error()
	if strings.HasPrefix(msg, "unknown flag:") ||
		strings.HasPrefix(msg, "unknown shorthand flag:") ||
		strings.HasPrefix(msg, "invalid argument") ||
		strings.HasPrefix(msg, "flag needs an argument") ||
		strings.HasPrefix(msg, "required flag(s)") {
		return exitcode.BadArgs
	}
	return exitcode.GeneralError
}

func buildRootCmd(ver, buildCommit string) *cobra.Command {
	root := &cobra.Command{
		Use:           "diskwhy",
		Short:         "Your disk is full. But why?",
		Long:          "diskwhy scans your disk and identifies what is consuming space, with developer-specific category awareness and safe cleanup.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       fmt.Sprintf("%s (build %s)", ver, buildCommit),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig(cmd)
		},
	}

	root.SetVersionTemplate("diskwhy {{.Version}}\n")

	root.PersistentFlags().Bool("no-color", false, "Disable color output")
	root.PersistentFlags().Bool("json", false, "Output as JSON (schema_version: 1)")
	root.PersistentFlags().Bool("verbose", false, "Show per-file timing, resolved paths, and diagnostic info")
	root.PersistentFlags().Bool("debug", false, "Show internal categorization decisions and Docker API responses")

	root.AddCommand(scanCmd)
	root.AddCommand(cleanCmd)

	return root
}

func initConfig(cmd *cobra.Command) error {
	config.BindFlags(cmd.Root().PersistentFlags())
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	GlobalConfig = cfg
	return nil
}
