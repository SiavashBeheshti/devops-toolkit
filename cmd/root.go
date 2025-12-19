package cmd

import (
	"os"

	"github.com/beheshti/devops-toolkit/cmd/compliance"
	"github.com/beheshti/devops-toolkit/cmd/docker"
	"github.com/beheshti/devops-toolkit/cmd/gitlab"
	"github.com/beheshti/devops-toolkit/cmd/k8s"
	"github.com/beheshti/devops-toolkit/pkg/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	version = "0.1.0"
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "devops-toolkit",
	Short: "A powerful DevOps CLI toolkit",
	Long: `DevOps Toolkit - A beautiful and powerful CLI for DevOps operations

Features:
  • Kubernetes operations (health checks, debugging, cleanup)
  • Docker container management and analysis
  • GitLab CI/CD pipeline management
  • Compliance and security checking

Examples:
  devops-toolkit k8s health          Check Kubernetes cluster health
  devops-toolkit docker stats        Show container statistics
  devops-toolkit gitlab pipelines    List GitLab pipelines
  devops-toolkit compliance check    Run compliance checks`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Show banner only for root command without subcommands
		if cmd.Name() == "devops-toolkit" && len(args) == 0 {
			output.Banner("DevOps Toolkit", "v"+version, "A powerful CLI for DevOps operations")
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.devops-toolkit.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringP("output", "o", "table", "output format (table, json, yaml)")

	// Bind flags to viper
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	_ = viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))

	// Add subcommands
	rootCmd.AddCommand(k8s.NewK8sCmd())
	rootCmd.AddCommand(docker.NewDockerCmd())
	rootCmd.AddCommand(gitlab.NewGitLabCmd())
	rootCmd.AddCommand(compliance.NewComplianceCmd())
	rootCmd.AddCommand(versionCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".devops-toolkit")
	}

	viper.SetEnvPrefix("DEVOPS")
	viper.AutomaticEnv()

	// Read config file if it exists (ignore error if config file doesn't exist)
	_ = viper.ReadInConfig()
}

// versionCmd shows version information
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		output.Header("DevOps Toolkit")
		output.Printf("  Version:    %s\n", version)
		output.Printf("  Go version: %s\n", "go1.21")
		output.Printf("  Platform:   %s/%s\n", "linux", "amd64")
		output.Newline()
	},
}

