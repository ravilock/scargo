package scraper

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/ravilock/scargo/internal/queue"
	"golang.org/x/net/html"
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
	// DomainRestriction represents the restriction imposed on the domains of the links when scraping, it has 3 different possiblities:
	//  - None: Scrape won't impose any restriction on domain names when scraping pages
	//  - SameDomainOnly: Scrape will only search pages that are in the same domain as the <url> argument
	//  - ListOfDomains: Scrape will only search pages that are passed on the DomainsList argument
	// Default value is SameDomainOnly
	DomainRestriction DomainRestriction

	// DomainList represents the list of domains that the Scrape function will be limited to when scraping pages
	// If DomainList is passed without "ListOfDomains" DomainRestriction configuration, this option will be ignored
	DomainList []string

	// DepthLimit represents the limit that the Scrape function will use when scraping some page
	// If DepthLimit is passed as 0, no limit will be used and the Scrape function will only stop if there are no more links to follow
	DepthLimit int
}

type depthedUrl struct {
	url   *url.URL
	depth int
}

func Scrape(starterUrl *url.URL, opts ...*ScrapeOptions) error {
	options := mergeScrapeOptions(opts)
	fmt.Println(starterUrl, options.DepthLimit, starterUrl.Hostname(), starterUrl.Path, options.DomainRestriction)
	urlQueue := queue.Queue[depthedUrl]{}
	urlQueue.Enqueue(depthedUrl{starterUrl, 0})
	for urlQueue.Length() != 0 {
		currentDepthedUrl, ok := urlQueue.Dequeue()
		if !ok {
			break
		}
		if currentDepthedUrl.depth > options.DepthLimit {
			break
		}
		currentUrl := currentDepthedUrl.url
		fmt.Printf("Current URL is %s and Depth is %d\n", currentDepthedUrl.url, currentDepthedUrl.depth)
		response, err := fetchPage(currentUrl)
		if err != nil {
			// TODO: Better handle this error to avoid stop scraping
			fmt.Println(err)
			continue
		}
		node, err := html.Parse(response)
		if err != nil {
			// TODO: Better handle this error to avoid stop scraping
			fmt.Println(err)
			continue
		}
		response.Close()
		scrapedUrls := scrapePageUrls(node)
		addHostNameToUrls(scrapedUrls, currentUrl)
		switch options.DomainRestriction {
		case ListOfDomains:
			scrapedUrls = filterDomainAndSchemes(scrapedUrls, options.DomainList)
		case None:
		case SameDomainOnly:
			scrapedUrls = filterDomainAndSchemes(scrapedUrls, []string{currentUrl.Hostname()})
		default:
			panic(fmt.Sprintf("unexpected scraper.DomainRestriction: %#v", options.DomainRestriction))
		}
		for _, currentUrl := range scrapedUrls {
			urlQueue.Enqueue(depthedUrl{currentUrl, currentDepthedUrl.depth + 1})
		}
	}
	return nil
}

func mergeScrapeOptions(opts []*ScrapeOptions) *ScrapeOptions {
	options := &ScrapeOptions{}
	for _, opt := range opts {
		if opt == nil {
			return options
		}
		return opt
	}
	return options
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

func scrapePageUrls(node *html.Node) []*url.URL {
	urls := []*url.URL{}
	if node.Type == html.ElementNode {
		for _, a := range node.Attr {
			if a.Key == "href" {
				parsedUrl, err := url.Parse(a.Val)
				if err != nil {
					break
				}
				urls = append(urls, parsedUrl)
				break
			}
		}
	}
	for childNode := node.FirstChild; childNode != nil; childNode = childNode.NextSibling {
		childUrls := scrapePageUrls(childNode)
		urls = append(urls, childUrls...)
	}
	return urls
}

func addHostNameToUrls(urls []*url.URL, originalUrl *url.URL) {
	for _, currentUrl := range urls {
		if currentUrl.Hostname() == "" {
			currentUrl.Host = originalUrl.Host
		}
		if currentUrl.Scheme == "" {
			currentUrl.Scheme = originalUrl.Scheme
		}
	}
}

func shouldFilterDomain(domainRestriction DomainRestriction) bool {
	return domainRestriction != None
}

func filterDomainAndSchemes(urls []*url.URL, domainList []string) []*url.URL {
	filteredUrls := make([]*url.URL, 0, len(urls))
	for _, currentUrl := range urls {
		if currentUrl.Scheme != "https" {
			continue
		}
		if currentUrl.Hostname() != "" && slices.Contains(domainList, currentUrl.Hostname()) {
			filteredUrls = append(filteredUrls, currentUrl)
		}
	}
	filteredUrls = slices.Clip(filteredUrls)
	return filteredUrls
}
