package scrape

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	flags := ScrapeCmd.Flags()

	flags.BoolVarP(&recursive, "recursive", "r", false, "Recursively scrape through all pages")
	flags.IntVarP(&depth, "depth", "d", 1, "Define sub-page depth limit to scrape pages")

	ScrapeCmd.MarkFlagsMutuallyExclusive("recursive", "depth")
}

var ScrapeCmd = &cobra.Command{
	Use:   "scrape <url> [-d Depth | -r]",
	Short: "Scrape pages from a root node",
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		fmt.Println("scrape", url, recursive, depth)
	},
	Args: cobra.ExactArgs(1),
}

var recursive bool
var depth int
