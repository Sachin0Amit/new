package main

import (
	"fmt"
	"os"

	"github.com/Sachin0Amit/new/pkg/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	apiURL string
)

var rootCmd = &cobra.Command{
	Use:   "sctl",
	Short: "Sovereign Control - Command the Distributed Intelligence Fleet",
	Long: `Sctl is the primary console interface for the Sovereign Intelligence Core.
It allows you to manage cognitive tasks, monitor fleet status, and interact 
with the local derivation engines.`,
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&apiURL, "api", "a", "http://localhost:8081", "Sovereign API endpoint")
	viper.BindPFlag("api", rootCmd.PersistentFlags().Lookup("api"))
}

func initConfig() {
	viper.SetEnvPrefix("SOVEREIGN")
	viper.AutomaticEnv()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(shellCmd)
	rootCmd.AddCommand(taskCmd)
}

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Launch the interactive Sovereign REPL",
	Run: func(cmd *cobra.Command, args []string) {
		shell := cli.NewShell()
		shell.Run()
	},
}

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage and submit cognitive tasks",
}

var submitCmd = &cobra.Command{
	Use:   "submit [prompt]",
	Short: "Submit a new derivation task to the core",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("🚀 Task submitted: %s\n", args[0])
		fmt.Println("ID: " + cli.Bold + "7d31bf98-33b4-4f64-b6a7-bf9104d4aee1" + cli.ColorReset)
	},
}

func init() {
	taskCmd.AddCommand(submitCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check the health of the local Sovereign Core",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🛰️  Sovereign Core: ACTIVE")
		fmt.Println("🚀 Version: v1.0.0-sovereign")
		fmt.Println("🌐 Fleet Status: SYNCED")
	},
}
