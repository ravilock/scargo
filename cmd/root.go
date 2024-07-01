package cmd

import (
	"log"

	"github.com/ravilock/scargo/cmd/exit"
	"github.com/ravilock/scargo/cmd/scrape"
	"github.com/ravilock/scargo/cmd/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "scargo [subcommand] [sub-subcommand]",
	Short: "Scargo Web Scrapper Cli",
	Long:  "", // TODO: Create long doc
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		panic(&exit.PanicError{Code: 1})
	}
}

func init() {
	commands := []*cobra.Command{
		version.VersionCmd,
		scrape.ScrapeCmd,
	}

	for _, cmd := range commands {
		rootCmd.AddCommand(cmd)
	}
}
