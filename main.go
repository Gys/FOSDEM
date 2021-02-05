package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func getSchedule() {
	res, err := http.Get("https://fosdem.org/2021/schedule/events/")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Find the events
	doc.Find("div#main table.table.table-striped.table-bordered.table-condensed tbody").Each(func(i int, s *goquery.Selection) {
		fmt.Printf("%s\n", s.Text())
	})
}

func main() {
	getSchedule()
}
