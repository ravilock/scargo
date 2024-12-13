package scraper

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/ravilock/scargo/internal/queue"
	urlparser "github.com/ravilock/scargo/urlParser"
)

type DomainRestriction int8

const (
	None DomainRestriction = iota - 1
	SameDomainOnly
	ListOfDomains
)

var httpClient = http.Client{}

// ScrapeOptions represent the options that can be passed to the Scrape function, it can be freely instatiated and passed by callers
type ScrapeOptions struct {
	// DomainList represents the list of domains that the Scrape function will be limited to when scraping pages
	// If DomainList is passed without "ListOfDomains" DomainRestriction configuration, this option will be ignored
	DomainList []string

	// DepthLimit represents the limit that the Scrape function will use when scraping some page
	// If DepthLimit is passed as 0, no limit will be used and the Scrape function will only stop if there are no more links to follow
	DepthLimit int

	// WorkerPoolSize represents the max size of the worker pool that is used to parallelize scraping and page fetching
	WorkerPoolSize int

	// DomainRestriction represents the restriction imposed on the domains of the links when scraping, it has 3 different possiblities:
	//  - None: Scrape won't impose any restriction on domain names when scraping pages
	//  - SameDomainOnly: Scrape will only search pages that are in the same domain as the <url> argument
	//  - ListOfDomains: Scrape will only search pages that are passed on the DomainsList argument
	// Default value is SameDomainOnly
	DomainRestriction DomainRestriction
}

type Scraper struct {
	Options *ScrapeOptions
	urlChan chan *DepthedUrl
	queue   queue.Queue[DepthedUrl]
}

type DepthedUrl struct {
	URL   *url.URL
	Depth int
}

func NewScraper(opts ...*ScrapeOptions) *Scraper {
	options := mergeScrapeOptions(opts)
	scraper := Scraper{
		Options: options,
		urlChan: make(chan *DepthedUrl),
	}
	return &scraper
}

func (s *Scraper) Start(startUrl *url.URL) {
	fmt.Println(startUrl, s.Options.DepthLimit, startUrl.Hostname(), startUrl.Path, s.Options.DomainRestriction)
	urlQueue := queue.Queue[*DepthedUrl]{}
	urlQueue.Enqueue(&DepthedUrl{startUrl, 0})
	for urlQueue.Length() != 0 {
		currentDepthedUrl, ok := urlQueue.Dequeue()
		if !ok {
			break
		}
		if currentDepthedUrl.Depth > s.Options.DepthLimit {
			break
		}
		scrapedUrls, err := s.scrape(currentDepthedUrl)
		if err != nil {
			fmt.Println(err)
			continue
		}
		for _, currentUrl := range scrapedUrls {
			urlQueue.Enqueue(currentUrl)
		}
	}
}

func (s *Scraper) scrape(currentUrl *DepthedUrl) ([]*DepthedUrl, error) {
	fmt.Printf("Current URL is %s and Depth is %d\n", currentUrl.URL, currentUrl.Depth)
	response, err := fetchPage(currentUrl.URL)
	if err != nil {
		// TODO: Better handle this error to avoid stop scraping
		return nil, err
	}
	urlParser := urlparser.UrlParser{OriginalURL: currentUrl.URL}
	switch s.Options.DomainRestriction {
	case ListOfDomains:
		urlParser.AllowedDomains = s.Options.DomainList
	case None:
	case SameDomainOnly:
		urlParser.AllowedDomains = []string{currentUrl.URL.Hostname()}
	default:
		panic(fmt.Sprintf("unexpected scraper.DomainRestriction: %#v", s.Options.DomainRestriction))
	}
	scrapedUrls, err := urlParser.ParseURLs(response)
	if err != nil {
		// TODO: Better handle this error to avoid stop scraping
		return nil, err
	}
	response.Close()
	depthedScrapedUrls := make([]*DepthedUrl, 0, len(scrapedUrls))
	for _, currentScrapedUrl := range scrapedUrls {
		depthedScrapedUrls = append(depthedScrapedUrls, &DepthedUrl{currentScrapedUrl, currentUrl.Depth + 1})
	}
	return depthedScrapedUrls, nil
}

func mergeScrapeOptions(opts []*ScrapeOptions) *ScrapeOptions {
	options := &ScrapeOptions{}
	if len(opts) == 0 || opts[0] == nil {
		return options
	}
	return opts[0]
}

func fetchPage(pageUrl *url.URL) (io.ReadCloser, error) {
	urlRequest, err := http.NewRequest("GET", pageUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	response, err := httpClient.Do(urlRequest)
	if err != nil {
		return nil, err
	}
	contentType := response.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		return nil, fmt.Errorf("invalid content type for scraper. Content type: %s", contentType)
	}
	return response.Body, nil
}
