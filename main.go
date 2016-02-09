package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

var s *StatusResp

func main() {
	log.Println("Authenticating...")
	w, err := NewWebhelper()
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Starting ticker...")
	go runStatusTicker(w)

	http.HandleFunc("/", statusHandler)
	log.Println("Starting server (:8080)...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Header.Get("Accept") {
	default:
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, formatStatusText(s))
	case "text/html":
		w.Header().Set("Content-Type", "text/html")
		io.WriteString(w, formatStatusHTML(s))
	}
}

func runStatusTicker(w *Webhelper) {
	c := time.Tick(1 * time.Second)
	for _ = range c {
		status, err := w.Status()
		if err != nil {
			log.Fatalln(err)
		}
		s = status
	}
}

func formatStatusText(s *StatusResp) string {
	if s == nil || !s.Running {
		return ""
	}

	artist := s.Track.ArtistResource.Name
	track := s.Track.TrackResource.Name
	album := s.Track.AlbumResource.Name
	url := s.Track.TrackResource.Location.OG
	duration := humanize(s.Track.Length)
	position := humanize(int(s.PlayingPosition))

	return fmt.Sprintf("[%s/%s] %s - %s (%s)\n%s\n",
		position, duration, artist, track, album, url)
}

func formatStatusHTML(s *StatusResp) string {
	if s == nil || !s.Running {
		return ""
	}

	artist := s.Track.ArtistResource.Name
	track := s.Track.TrackResource.Name
	album := s.Track.AlbumResource.Name
	url := s.Track.TrackResource.Location.OG
	duration := humanize(s.Track.Length)
	position := humanize(int(s.PlayingPosition))

	return fmt.Sprintf(`
<html>
<head>
<title>[%[1]s/%[2]s] %[3]s - %[4]s (%[5]s)</title>
</head>
<body>
<div>[%[1]s/%[2]s] <a href="%[3]s">%[4]s - %[5]s (%[6]s)</a><div>
</body>
</html>
`, position, duration, artist, track, album, url)
}

func humanize(duration int) string {
	minutes := duration / 60
	seconds := duration % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}
