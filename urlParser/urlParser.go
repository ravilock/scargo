package urlparser

import (
	"fmt"
	"io"
	"net/url"
	"slices"

	"golang.org/x/net/html"
)

type UrlParser struct {
	OriginalURL    *url.URL
	AllowedDomains []string
}

func (u *UrlParser) ParseURLs(r io.ReadCloser) ([]*url.URL, error) {
	node, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse page: %w", err)
	}
	return u.scrapePageUrls(node), nil
}

func (u *UrlParser) scrapePageUrls(node *html.Node) []*url.URL {
	urls := []*url.URL{}
	if node.Type == html.ElementNode {
		for _, a := range node.Attr {
			if a.Key == "href" {
				parsedUrl, err := url.Parse(a.Val)
				if err != nil {
					break
				}
				parsedUrl = u.addHostNameToUrls(parsedUrl)
				if u.isAllowedDomain(parsedUrl) {
					urls = append(urls, parsedUrl)
				}
				break
			}
		}
	}
	for childNode := node.FirstChild; childNode != nil; childNode = childNode.NextSibling {
		childUrls := u.scrapePageUrls(childNode)
		urls = append(urls, childUrls...)
	}
	return urls
}

func (u *UrlParser) isAllowedDomain(currentUrl *url.URL) bool {
	if currentUrl.Scheme != "https" {
		return false
	}
	if currentUrl.Hostname() != "" && slices.Contains(u.AllowedDomains, currentUrl.Hostname()) {
		return true
	}
	return false
}

func (u *UrlParser) addHostNameToUrls(currentUrl *url.URL) *url.URL {
	if currentUrl.Hostname() == "" {
		currentUrl.Host = u.OriginalURL.Host
	}
	if currentUrl.Scheme == "" {
		currentUrl.Scheme = u.OriginalURL.Scheme
	}
	return currentUrl
}
