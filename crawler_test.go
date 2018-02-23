package main

import (
	"testing"
	"io"
	"bytes"
)

func TestCorrectHostDiffScheme(t *testing.T) {
	c := &Crawler{RootURL: "http://test.go"}
	c.init()
	url := "https://test.go"
	isCorrect := c.isCorrectHost(url)
	if !isCorrect {
		t.Errorf("URL isn't considered to have a correct host: %q", url)
	}
}

func TestCorrectHostRelative(t *testing.T) {
	c := &Crawler{RootURL: "http://test.go"}
	c.init()
	url := "http://test.go/some/link"
	isCorrect := c.isCorrectHost(url)
	if !isCorrect {
		t.Errorf("URL isn't considered to have a correct host: %q", url)
	}
}

func TestRelativeLink(t *testing.T) {
	c := &Crawler{RootURL: "http://test.go"}
	c.init()
	res, err := c.getLink("/test")
	if err != nil {
		t.Errorf("Relative url not parsed: %q", err)
	}
	if res != "http://test.go/test" {
		t.Errorf("Relative url not parsed: %q", res)
	}
}

func TestAbsoluteLink(t *testing.T) {
	c := &Crawler{RootURL: "http://test.go"}
	c.init()
	res, err := c.getLink("http://test.go/2")
	if err != nil {
		t.Errorf("Absolute url not parsed: %q", err)
	}
	if res != "http://test.go/2" {
		t.Errorf("Absolute url not parsed: %q", res)
	}
}

func TestAbsoluteExternalLink(t *testing.T) {
	c := &Crawler{RootURL: "http://test.go"}
	c.init()
	_, err := c.getLink("http://test2.go")
	if err == nil {
		t.Errorf("External url incorrectly parsed: %q", err)
	}
}

func TestGetLinksSingle(t *testing.T) {
	c := &Crawler{RootURL: "http://test.go"}
	c.init()
	body := io.Reader(bytes.NewBuffer([]byte("<a href='http://test.go'>")))
	links := c.getLinks(body)
	if _, ok := links["http://test.go"]; !ok || len(links) != 1 {
		t.Errorf("HTML parsed incorrectly: %v", links)
	}
}

func TestGetLinksAnchor(t *testing.T) {
	c := &Crawler{RootURL: "http://test.go"}
	c.init()
	body := io.Reader(bytes.NewBuffer([]byte("<a href='http://test.go#123'>")))
	links := c.getLinks(body)
	if _, ok := links["http://test.go"]; !ok || len(links) != 1 {
		t.Errorf("HTML parsed incorrectly: %v", links)
	}
}

func TestGetLinksExternal(t *testing.T) {
	c := &Crawler{RootURL: "http://test.go"}
	c.init()
	body := io.Reader(bytes.NewBuffer([]byte("<a href='http://test2.go'>")))
	links := c.getLinks(body)
	if len(links) != 0 {
		t.Errorf("HTML parsed incorrectly: %v", links)
	}
}

func TestGetLinksMultiple(t *testing.T) {
	c := &Crawler{RootURL: "http://test.go"}
	c.init()
	body := io.Reader(bytes.NewBuffer([]byte("<a href='http://test.go'><img/><a href='http://test.go/link'>")))
	links := c.getLinks(body)
	if len(links) != 2 {
		t.Errorf("HTML parsed incorrectly: %v", links)
	}
}