package main

import (
	"sync"
	"net/url"
	"log"
	"os"
	"net/http"
	"io"
	"golang.org/x/net/html"
	"strings"
	"regexp"
	"errors"
)

type Crawler struct {
	RootURL       string
	log           *log.Logger
	parsedRootURL *url.URL
	host          string
	lock          sync.Mutex
	wg            *sync.WaitGroup
	visited       map[string]interface{}
	httpClient    *http.Client
	anchorFilter  *regexp.Regexp
}

func (c *Crawler) doRequest(method string, urlStr string) (*http.Response, error) {
	req, err := http.NewRequest(method, urlStr, nil)
	if err != nil {
		c.log.Printf("Failed to init request to URL %q", urlStr)
		return nil, err
	}
	req.Header.Set("User-Agent", "monzo-web-crawler")
	res, err := c.httpClient.Do(req)
	if err != nil {
		c.log.Printf("Error sending request to URL %q", urlStr)
		return nil, err
	}
	return res, nil
}

func (c *Crawler) isHTML(urlStr string) (bool, error) {
	resp, err := c.doRequest("HEAD", urlStr)
	if err != nil {
		return false, err
	}
	if r := resp.Header.Get("Content-Type"); !strings.HasPrefix(r, "text/html") {
		return false, nil
	}
	return true, nil
}

func (c *Crawler) getHTML(urlStr string) (*http.Response, error) {
	return c.doRequest("GET", urlStr)
}

func (c *Crawler) setVisited(key string, val interface{}) {
	c.lock.Lock()
	c.visited[key] = val
	c.lock.Unlock()
}

func (c *Crawler) crawl(urlStr string) {
	defer c.wg.Done()

	// Check if any other goroutine has already processed this URL
	c.lock.Lock()
	_, visited := c.visited[urlStr]
	if visited {
		c.lock.Unlock()
		return
	} else {
		// Set preliminary status to mark ownership of this URL
		c.visited[urlStr] = "visited"
		c.lock.Unlock()
	}

	isHTML, err := c.isHTML(urlStr)
	if err != nil {
		c.setVisited(urlStr, "error")
		return
	}

	if !isHTML {
		// js maybe? skip anyway
		c.setVisited(urlStr, "non-html")
		return
	}

	resp, err := c.getHTML(urlStr)
	if err != nil {
		c.setVisited(urlStr, "error")
		return
	}

	links := c.getLinks(resp.Body)
	resp.Body.Close()

	c.setVisited(urlStr, links)
	for link := range links {
		c.wg.Add(1)
		go c.crawl(link)
	}
}

func (c *Crawler) parseURL(urlStr string) (*url.URL, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		c.log.Printf("Cannot parse URL %q", urlStr)
		return nil, err
	}
	return parsedURL, nil
}

func (c *Crawler) isCorrectHost(urlStr string) bool {
	parsedURL, err := c.parseURL(urlStr)
	if err != nil {
		return false
	}
	if parsedURL.Hostname() != c.host {
		return false
	}
	return true
}

func (c *Crawler) getLinks(body io.Reader) map[string]bool {
	links := make(map[string]bool) // Using map as a HashSet
	tokenizer := html.NewTokenizer(body)
	for {
		elem := tokenizer.Next()
		switch elem {
		case html.ErrorToken:
			return links
		case html.StartTagToken, html.EndTagToken:
			token := tokenizer.Token()
			if token.Data != "a" {
				continue
			}
			for _, attr := range token.Attr {
				if attr.Key != "href" {
					continue
				}
				link, err := c.getLink(attr.Val)
				if err == nil {
					links[link] = true
				}
			}
		}
	}
}

func (c *Crawler) getLink(rawLink string) (link string, err error) {
	if strings.HasPrefix(rawLink, "/") {
		// Transform relative link into absolute
		link = (&url.URL{Scheme: c.parsedRootURL.Scheme, Host: c.host, Path: rawLink}).String()
	} else if c.isCorrectHost(rawLink) {
		link = rawLink
	} else {
		return "", errors.New("link to an external host")
	}
	// Strip anchor part from the end of the link
	link = c.anchorFilter.ReplaceAllLiteralString(link, "")
	return link, nil
}

func (c *Crawler) init() error {
	var err error
	c.parsedRootURL, err = c.parseURL(c.RootURL)
	if err != nil {
		return err
	}
	c.anchorFilter, err = regexp.Compile(`(%23|#)[\w\d\-]+$`) // %23 is urlencoded #
	if err != nil {
		return err
	}

	c.log = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
	c.httpClient = &http.Client{}
	c.host = c.parsedRootURL.Hostname()
	c.visited = make(map[string]interface{})
	c.wg = &sync.WaitGroup{}
	return nil
}

func (c *Crawler) Run() (*map[string]interface{}, error) {
	err := c.init()
	if err != nil {
		return nil, err
	}

	c.wg.Add(1)
	go c.crawl(c.RootURL)
	c.wg.Wait()

	return &c.visited, nil
}
