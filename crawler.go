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
)

type Crawler struct {
	RootURL       string
	log           *log.Logger
	parsedRootURL *url.URL
	host          string
	lock          sync.Mutex
	wg            *sync.WaitGroup
	visited       map[string]interface{}
	HttpClient    *http.Client
}

func (c *Crawler) doRequest(urlStr string) (*http.Response, error) {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		c.log.Printf("Failed to init request to URL %q", urlStr)
		return nil, err
	}
	req.Header.Set("User-Agent", "monzo-web-crawler")
	res, err := c.HttpClient.Do(req)
	if err != nil {
		c.log.Printf("Error sending request to URL %q", urlStr)
		return nil, err
	}
	return res, nil
}

func (c *Crawler) crawl(urlStr string) {
	defer c.wg.Done()
	c.lock.Lock()
	_, visited := c.visited[urlStr]
	if visited {
		c.lock.Unlock()
		return
	} else {
		c.visited[urlStr] = "visited"
		c.lock.Unlock()
	}

	resp, err := c.doRequest(urlStr)
	if err != nil {
		c.lock.Lock()
		c.visited[urlStr] = "error"
		c.lock.Unlock()
		return
	}
	links := c.getLinks(resp.Body)
	resp.Body.Close()

	c.lock.Lock()
	c.visited[urlStr] = links
	c.lock.Unlock()

	for link := range links {
		c.wg.Add(1)
		go c.crawl(link)
	}
}

func (c *Crawler) isCorrectHost(urlStr string) bool {
	parsedURL, err := c.parseURL(urlStr)
	if err != nil {
		return false
	}
	if parsedURL.Hostname() != c.host {
		//c.log.Printf("External host, not following: %q", urlStr)
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
				var link string
				if strings.HasPrefix(attr.Val, "/") {
					link = (&url.URL{Scheme: c.parsedRootURL.Scheme, Host: c.host, Path: attr.Val}).String()
				} else if c.isCorrectHost(attr.Val) {
					link = attr.Val
				} else {
					continue
				}
				if strings.HasSuffix(link, ".jpg") || strings.HasSuffix(link, ".png") {
					continue
				}
				links[link] = true
			}
		}
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

func (c *Crawler) Start() (*map[string]interface{}, error) {
	var err error
	c.parsedRootURL, err = c.parseURL(c.RootURL)
	if err != nil {
		return nil, err
	}
	c.log = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
	c.HttpClient = &http.Client{}
	c.host = c.parsedRootURL.Hostname()
	c.visited = make(map[string]interface{})
	c.wg = &sync.WaitGroup{}
	c.wg.Add(1)
	go c.crawl(c.RootURL)
	c.wg.Wait()
	return &c.visited, nil
}
