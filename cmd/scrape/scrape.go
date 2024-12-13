package scrape

import (
	"fmt"
	"log"
	"net/url"
	"runtime"

	"github.com/ravilock/scargo/cmd/exit"
	"github.com/ravilock/scargo/scraper"
	"github.com/spf13/cobra"
)

var (
	depth              int
	domainRestriction  int8
	numberOfGoroutines int
)

func init() {
	flags := ScrapeCmd.Flags()

	flags.IntVarP(&depth, "depth", "d", 5, "Define sub-page depth limit to scrape pages. Set to 0 for no limit.")
	flags.Int8Var(&domainRestriction, "domain-restriction", 0, "Define the domain restriction imposed on the Scrape function.")
	flags.IntVar(&numberOfGoroutines, "worker-pool-size", runtime.NumCPU(), "Define the worker pool size that is going to be used on the Scraper.")
}

var ScrapeCmd = &cobra.Command{
	Use:   "scrape <url> [-d Depth] [--domain-restriction DomainRestriction]",
	Short: "Scrape pages from a root node",
	Run:   scrape,
	Args:  cobra.ExactArgs(1),
}

func scrape(cmd *cobra.Command, args []string) {
	urlArg := args[0]
	parsedUrl, err := url.ParseRequestURI(urlArg)
	if err != nil {
		log.Println("Failed to parse URL", err)
		panic(&exit.PanicError{Code: 1})
	}
	webScraper := scraper.NewScraper(&scraper.ScrapeOptions{
		DepthLimit:        depth,
		WorkerPoolSize:    numberOfGoroutines,
		DomainRestriction: scraper.DomainRestriction(domainRestriction),
	})
	webScraper.Start(parsedUrl)
	fmt.Println("end")
}
