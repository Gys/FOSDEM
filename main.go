package main

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type eventDetails struct {
	titleHTML       string
	speakersHTML    string
	roomHTML        string
	start           time.Time
	end             time.Time
	attachmentsHTML string
	videoHTML       string
}

func getSchedule() (list []eventDetails) {
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

	// Find the records with events
	tr := doc.Find("div#main table.table.table-striped.table-bordered.table-condensed tbody").Children()
	tr.Each(func(i int, s *goquery.Selection) {
		// Get the data
		dt := s.Children()
		html, err := dt.First().Html()
		// Skip records having room name
		if err != nil || strings.HasPrefix(html, "<h4>") {
			return
		}
		event := eventDetails{}
		dt.Each(func(j int, e *goquery.Selection) {
			// Title with link
			if j == 0 {
				event.titleHTML = htmlWithFullUrls(e)
				// Or maybe split text and link
				// event.title = e.Text()
				// link, _ := e.Find("a").Attr("href")
				// event.link = "https://fosdem.org" + link
			}
			// Speakers
			if j == 1 {
				event.speakersHTML = htmlWithFullUrls(e)
			}
			// Room
			if j == 2 {
				event.roomHTML = htmlWithFullUrls(e)
			}
			// Day
			if j == 3 {
				v := "2021-02-06 WET" // Saturday
				if e.Text() == "Sunday" {
					v = "2021-02-07 WET"
				}
				event.start, _ = time.Parse("2006-01-02 MST", v)
				event.end = event.start
			}
			// Start
			if j == 4 {
				d := e.Text()
				if len(d) == 5 {
					d = strings.Replace(d, ":", "h", -1)
					d += "m"
					dd, _ := time.ParseDuration(d)
					event.start = event.start.Add(dd)
				}
			}
			// End
			if j == 5 {
				d := e.Text()
				if len(d) == 5 {
					d = strings.Replace(d, ":", "h", -1)
					d += "m"
					dd, _ := time.ParseDuration(d)
					event.end = event.end.Add(dd)
				}
			}
			// Attachments
			if j == 6 {
				event.attachmentsHTML = htmlWithFullUrls(e)
			}
			// Video
			if j == 7 {
				event.videoHTML = htmlWithFullUrls(e)
			}
		})
		list = append(list, event)
	})
	return
}

func main() {
	list := getSchedule()
	sort.Slice(list, func(i int, j int) bool {
		return list[i].start.Before(list[j].start)
	})
	for _, event := range list {
		fmt.Printf("%s %s\n", event.titleHTML, event.start)
	}
}

func htmlWithFullUrls(s *goquery.Selection) string {
	html, err := s.Html()
	if err != nil {
		return ""
	}
	html = strings.ReplaceAll(html, "href=\"", "href=\"https://fosdem.org")
	return html
}
