/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gabriel-samfira/localshow/config"
	"github.com/gabriel-samfira/localshow/httpsrv"
	"github.com/gabriel-samfira/localshow/params"
	"github.com/gabriel-samfira/localshow/sshsrv"
	"github.com/spf13/cobra"
)

var (
	cfgFile string = "/etc/localshow/localshow.toml"
)

var signals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "localshowd",
	Short: "A simple HTTP(S) reverse proxy over ssh tunnel",
	Long:  ``,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, stop := signal.NotifyContext(context.Background(), signals...)
		defer stop()

		tunnelEvents := make(chan params.TunnelEvent, 100)

		cfg, err := config.NewConfig(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if err := sshsrv.GenerateKey(cfg.SSHServer.HostKeyPath); err != nil {
			return fmt.Errorf("failed to generate host key: %w", err)
		}

		sshSrv, err := sshsrv.NewSSHServer(ctx, cfg, tunnelEvents)
		if err != nil {
			return fmt.Errorf("failed to create ssh server: %w", err)
		}

		if err := sshSrv.Start(); err != nil {
			return fmt.Errorf("failed to start ssh server: %w", err)
		}

		httpSrv, err := httpsrv.NewHTTPServer(ctx, cfg, tunnelEvents)
		if err != nil {
			return fmt.Errorf("failed to create http server: %w", err)
		}

		if err := httpSrv.Start(); err != nil {
			return fmt.Errorf("failed to start http server: %w", err)
		}

		<-ctx.Done()
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", cfgFile, "config file for localshowd")
}
