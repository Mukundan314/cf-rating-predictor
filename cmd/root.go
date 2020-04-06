package cmd

import "github.com/spf13/cobra"

var (
	rootCmd = &cobra.Command{
		Use:   "cf-rating-predictor [flags]",
		Short: "Server for cf-rating-predictor",
		Run:   run,
	}
)

func init() {
}

func run(cmd *cobra.Command, args []string) {
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
