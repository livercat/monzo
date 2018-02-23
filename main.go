package main

import "fmt"

func pprint(visitedLinks *map[string]interface{}) {
	for url, links := range *visitedLinks {
		fmt.Printf("\n%s\n", url)
		switch links.(type) {
		case map[string]bool:
			for link := range links.(map[string]bool) {
				fmt.Printf("-- %s\n", link)
			}
		case string:
			fmt.Printf("-- %s\n", links)
		}
	}
}

func main() {
	// TODO: support robots.txt
	// TODO: set a global crawl timeout using context
	// TODO: move url, user-agent, timeout, etc. into command-line flags
	// TODO: improve site map presentation
	crawler := Crawler{RootURL: "https://monzo.com"}
	res, err := crawler.Run()
	if err != nil {
		fmt.Print(err)
		panic(err)
	}
	pprint(res)
}
