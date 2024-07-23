package urlparser

import (
	"fmt"
	"io"
	"net/url"

	"golang.org/x/net/html"
)

type UrlParser struct {
	AllowedDomains []string
}

func (u *UrlParser) ParseURLs(r io.ReadCloser) ([]*url.URL, error) {
	node, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse page: %w", err)
	}
	return nil, nil
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
