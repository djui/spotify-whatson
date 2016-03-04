package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"text/template"
	"time"
)

type status struct {
	Artist   string
	Track    string
	Album    string
	URL      string
	Duration string
	Position string
}

var s *StatusResp
var templateHTML *template.Template
var templateText *template.Template

func init() {
	templateHTML = template.Must(template.New("html").Parse(templateHTMLFormat))
	templateText = template.Must(template.New("text").Parse(templateTextFormat))
}

func main() {
	port := flag.Int("port", 8080, "Port")

	flag.Parse()

	log.Println("Authenticating...")
	w, err := NewWebhelper()
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Starting ticker...")
	go runStatusTicker(w)

	address := fmt.Sprintf(":%d", *port)
	http.HandleFunc("/", statusHandler)
	log.Printf("Starting server (%s)...", address)
	log.Fatal(http.ListenAndServe(address, nil))
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	if s == nil || !s.Running {
		return
	}

	status := parseStatus(s)

	acceptType := strings.Split(r.Header.Get("Accept"), ",")[0]
	switch acceptType {
	default:
		w.Header().Set("Content-Type", "text/plain")
		templateText.Execute(w, status)
	case "text/html":
		w.Header().Set("Content-Type", "text/html")
		templateHTML.Execute(w, status)
	}
}

func runStatusTicker(w *Webhelper) {
	c := time.Tick(1 * time.Second)
	for _ = range c {
		status, err := w.Status()
		if err != nil {
			log.Println("Warning:", err)
		}
		s = status
	}
}

func parseStatus(s *StatusResp) *status {
	return &status{
		Artist:   s.Track.ArtistResource.Name,
		Track:    s.Track.TrackResource.Name,
		Album:    s.Track.AlbumResource.Name,
		URL:      s.Track.TrackResource.Location.OG,
		Duration: humanize(s.Track.Length),
		Position: humanize(int(s.PlayingPosition)),
	}
}

func humanize(duration int) string {
	minutes := duration / 60
	seconds := duration % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

const templateTextFormat = `[{{.Position}}/{{.Duration}}] {{.Artist}} - {{.Track}} ({{.Album}})
{{.URL}}
`
const templateHTMLFormat = `
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta http-equiv="refresh" content="1">
  <title>[{{.Position}}/{{.Duration}}] {{.Artist}} - {{.Track}} ({{.Album}})</title>
</head>
<body>
  <div>[{{.Position}}/{{.Duration}}] <a href="{{.URL}}">{{.Artist}} - {{.Track}}</a> ({{.Album}})<div>
</body>
</html>
`
