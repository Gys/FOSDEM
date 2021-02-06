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

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
)

type eventDetails struct {
	TitleHTML       string
	TitleText       string
	TitleLink       string
	SpeakersHTML    string
	RoomHTML        string
	Start           time.Time
	End             time.Time
	AttachmentsHTML string
	VideoHTML       string
	ID              string // unique id for each talk, use title url
}

func (e eventDetails) StartAsHTML() string {
	return e.Start.Format("15:04") + "&nbsp;-&nbsp;" + e.End.Format("15:04")
}

func main() {
	list := getSchedule()
	// Sort by datetime start
	// TODO: sort also by end time
	sort.Slice(list, func(i int, j int) bool {
		return list[i].Start.Before(list[j].Start)
	})
	writeHTML("fosdem_schedules.html", list)
	// writeMD("fosdem_schedules.md", list)
	// writeCSV("fosdem_schedules.csv", list)
}

func writeHTML(fn string, list []eventDetails) {
	f, err := os.Create(fn)
	if err != nil {
		log.Fatalf("Unable to create '%s' because %s", fn, err)
	}
	defer f.Close()
	t := template.Must(template.New("").Parse(htmlTemplate))
	if err := t.Execute(f, &list); err != nil {
		log.Fatal(err)
	}
}

func writeCSV(fn string, list []eventDetails) {
	// experimental
	// TODO: how to write a variable number of text/link combinations in a useful way?
	f, err := os.Create(fn)
	if err != nil {
		log.Fatalf("Unable to create '%s' because %s", fn, err)
	}
	defer f.Close()
	for i := range list {
		fmt.Fprintf(f, "\"%s\"", list[i].TitleText)
		fmt.Fprintf(f, ",\"%s\"", list[i].TitleLink)
		fmt.Fprintf(f, ",\"%s\"", list[i].RoomHTML)
		fmt.Fprintf(f, ",\"%s\"", list[i].Start.Format("2006-01-02 15:04"))
		fmt.Fprintf(f, ",\"%s\"", list[i].End.Format("2006-01-02 15:04"))
		fmt.Fprintf(f, "\n")
	}
}

func writeMD(fn string, list []eventDetails) {
	// experimental
	f, err := os.Create(fn)
	if err != nil {
		log.Fatalf("Unable to create '%s' because %s", fn, err)
	}
	defer f.Close()
	converter := md.NewConverter("", true, nil)
	for i := range list {
		list[i].TitleHTML, _ = converter.ConvertString(list[i].TitleHTML)
		list[i].SpeakersHTML, _ = converter.ConvertString(list[i].SpeakersHTML)
		list[i].RoomHTML, _ = converter.ConvertString(list[i].RoomHTML)
		list[i].AttachmentsHTML, _ = converter.ConvertString(list[i].AttachmentsHTML)
		list[i].VideoHTML, _ = converter.ConvertString(list[i].VideoHTML)
	}
	t := template.Must(template.New("").Parse(mdTemplate))
	if err := t.Execute(f, &list); err != nil {
		log.Fatal(err)
	}
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
			// Title
			if j == 0 {
				event.TitleHTML = htmlWithFullUrls(e)
				event.ID, _ = e.Find("a").Attr("href")
				t, l := splitHTML(e)
				event.TitleText = first(t)
				event.TitleLink = first(l)
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

func first(s []string) string {
	if len(s) == 0 {
		return ""
	}
	return s[0]
}

func splitHTML(s *goquery.Selection) (texts []string, links []string) {
	s.Children().Each(func(i int, c *goquery.Selection) {
		texts = append(texts, c.Text())
		link, _ := c.Attr("href")
		if !strings.HasPrefix(link, "https://fosdem.org") {
			link = "https://fosdem.org" + link
		}
		links = append(links, link)
	})
	return
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
	// Open all links in new tab
	html = strings.ReplaceAll(html, "href=\"", "target=\"_blank\" href=\"")
	return html
}

const htmlTemplate = `
<!doctype html>
<html lang="en">
	<head>
		<meta charset="utf-8">
		<title>The HTML5 Herald</title>
		<link media="all" rel="stylesheet" type="text/css" href="fosdem_schedules.css">
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
			<tr><td>{{.StartAsHTML}}</td><td>{{.RoomHTML}}</td><td><input type="checkbox" id="{{.ID}}" onclick="handleClick(event)"></td><td>{{.TitleHTML}}</td><td>{{.AttachmentsHTML}}</td><td>{{.SpeakersHTML}}</td><td>{{.VideoHTML}}</td></tr>
			{{end}}
			</table>
		</div>
		<script src="https://cdn.jsdelivr.net/npm/js-cookie@rc/dist/js.cookie.min.js"></script>
		<script>
			for (let box in Cookies.get()) {
				document.getElementById(box).checked = true;
			}
			function handleClick(e) {
				if (e.target.checked) {
					Cookies.set(e.target.id, true, { expires: 180 });
				} else {
					Cookies.remove(e.target.id);
				}
				console.log(Cookies.get());
			}
		</script>
	</body>
</html>
`

const mdTemplate = `
{{range .}}
{{.StartAsHTML}}|{{.RoomHTML}}|{{.TitleHTML}}|{{.AttachmentsHTML}}|{{.SpeakersHTML}}|{{.VideoHTML}}
{{end}}
`
