package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type eventDetails struct {
	TitleHTML       string
	SpeakersHTML    string
	RoomHTML        string
	Start           time.Time
	End             time.Time
	AttachmentsHTML string
	VideoHTML       string
}

func (e eventDetails) StartAsHTML() string {
	return e.Start.Format("15:04") + "&nbsp;-&nbsp;" + e.End.Format("15:04")
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
				event.TitleHTML = htmlWithFullUrls(e)
				// TODO: maybe split text and link
				// event.title = e.Text()
				// link, _ := e.Find("a").Attr("href")
				// event.link = "https://fosdem.org" + link
			}
			// Speakers
			if j == 1 {
				event.SpeakersHTML = htmlWithFullUrls(e)
			}
			// Room
			if j == 2 {
				event.RoomHTML = htmlWithFullUrls(e)
			}
			// Day
			if j == 3 {
				v := "2021-02-06 WET" // Saturday
				if e.Text() == "Sunday" {
					v = "2021-02-07 WET"
				}
				event.Start, _ = time.Parse("2006-01-02 MST", v)
				event.End = event.Start
			}
			// Start
			if j == 4 {
				d := e.Text()
				if len(d) == 5 {
					d = strings.Replace(d, ":", "h", -1)
					d += "m"
					dd, _ := time.ParseDuration(d)
					event.Start = event.Start.Add(dd)
				}
			}
			// End
			if j == 5 {
				d := e.Text()
				if len(d) == 5 {
					d = strings.Replace(d, ":", "h", -1)
					d += "m"
					dd, _ := time.ParseDuration(d)
					event.End = event.End.Add(dd)
				}
			}
			// Attachments
			if j == 6 {
				event.AttachmentsHTML = htmlWithFullUrls(e)
			}
			// Video
			if j == 7 {
				event.VideoHTML = htmlWithFullUrls(e)
			}
		})
		list = append(list, event)
	})
	return
}

func main() {
	list := getSchedule()
	// Sort by datetime start
	sort.Slice(list, func(i int, j int) bool {
		return list[i].Start.Before(list[j].Start)
	})
	for _, event := range list {
		fmt.Printf("%s %s\n", event.TitleHTML, event.Start)
	}
	fn := "fosdem_schedule.html"
	f, err := os.Create(fn)
	if err != nil {
		log.Fatalf("Unable to create '%s' because %s", fn, err)
	}
	defer f.Close()
	t := template.Must(template.New("").Parse(pageTemplate))
	if err := t.Execute(f, &list); err != nil {
		log.Fatal(err)
	}
}

func htmlWithFullUrls(s *goquery.Selection) string {
	html, err := s.Html()
	if err != nil {
		return ""
	}
	// Assumption: if one link without the domain then no links have a domain (because attachments seem to have a full link)
	if !strings.Contains(html, "https://fosdem.org") {
		html = strings.ReplaceAll(html, "href=\"", "href=\"https://fosdem.org")
	}
	return html
}

const pageTemplate = `
<!doctype html>
<html lang="en">
	<head>
		<meta charset="utf-8">
		<title>The HTML5 Herald</title>
		<link media="all" rel="stylesheet" type="text/css" href="https://fosdem.org/2021/assets/style/fosdem-18736d187ceb9d8deb0e21312ca92ecbafa3786eabacf5c3a574d0f73c273843.css">
		<style>
			#main {
				max-width: 100%;
			}
		</style>
	</head>
	<body class="schedule-events">
		<div id="main">
			<table class="table table-striped table-bordered table-condensed">
			{{range .}}
			<tr><td>{{.StartAsHTML}}</td><td>{{.RoomHTML}}</td><td>{{.TitleHTML}}</td><td>{{.AttachmentsHTML}}</td><td>{{.SpeakersHTML}}</td><td>{{.VideoHTML}}</td></tr>
			{{end}}
			</table>
		</div>
	</body>
</html>
`
