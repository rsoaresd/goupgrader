package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = NewRootCmd()

func NewRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "goupgrader",
		Short: "A command-line to upgrade Go project dependencies using config-based automation",
		Long: `goupgrader helps automate the process of upgrading Go project dependencies, 
particularly for projects targeting specific OpenShift version. It provides commands to 
generate config files based on OpenShift version dependencies and to upgrade dependencies 
accordingly using the generated configuration.`,
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(NewGenerateConfigForOpenshiftDependencies())
	rootCmd.AddCommand(NewUpgrade())
}
