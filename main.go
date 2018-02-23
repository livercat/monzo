package main

import "fmt"

func pprint(visitedLinks *map[string]interface{}) {
	for url, links := range *visitedLinks {
		fmt.Println(url)
		for link := range links.(map[string]bool) {
			fmt.Printf("-- %s\n", link)
		}
	}
}

func main() {
	crawler := Crawler{RootURL: "https://monzo.com"}
	res, _ := crawler.Start()
	pprint(res)
}
