package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"text/template"
	"time"

	"golang.org/x/net/websocket"
)

type status struct {
	Artist   string
	Track    string
	Album    string
	URL      string
	Duration string
	Position string
	Context  interface{} // Currently unused
}

var s *StatusResp
var templateHTML *template.Template
var templateText *template.Template
var templateWS *template.Template

func init() {
	templateHTML = template.Must(template.New("html").Parse(templateHTMLFormat))
	templateText = template.Must(template.New("text").Parse(templateTextFormat))
	templateWS = template.Must(template.New("ws").Parse(templateWSFormat))
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
	http.Handle("/ws", websocket.Handler(statusPushHandler))
	log.Printf("Starting server (%s)...", address)
	log.Fatal(http.ListenAndServe(address, nil))
}

func statusPushHandler(conn *websocket.Conn) {
	for _ = range time.Tick(1 * time.Second) {
		if s == nil || !s.Running {
			return
		}

		status := parseStatus(s)
		buf := new(bytes.Buffer)
		templateWS.Execute(buf, status)
		fmt.Fprintf(conn, "%s", buf)
	}
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	if s == nil || !s.Running {
		return
	}

	status := parseStatus(s)

	switch acceptType(r) {
	default:
		w.Header().Set("Content-Type", "text/plain")
		templateText.Execute(w, status)
	case "text/html":
		w.Header().Set("Content-Type", "text/html")
		templateHTML.Execute(w, status)
	}
}

func acceptType(r *http.Request) string {
	return strings.Split(r.Header.Get("Accept"), ",")[0]
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
		Context:  s.Context,
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

const templateWSFormat = `[{{.Position}}/{{.Duration}}] <a href="{{.URL}}">{{.Artist}} - {{.Track}}</a> ({{.Album}})`

const templateHTMLFormat = `
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>[{{.Position}}/{{.Duration}}] {{.Artist}} - {{.Track}} ({{.Album}})</title>
</head>
<body>
  <div id="status"></div>

	<script>
	webSocket = new WebSocket("ws://" + window.location.host + "/ws");
	webSocket.onmessage = function(event) {
		document.querySelector("title").innerHTML = event.data;
		document.querySelector("#status").innerHTML = event.data;
	};
	</script>
</body>
</html>
`
