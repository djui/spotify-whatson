package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
)

func init() {
	http.DefaultClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
}

type Webhelper struct {
	host   string
	port   int
	params url.Values
	header http.Header
}

func NewWebhelper() (*Webhelper, error) {
	host := generateRandomHostname()
	port := 4370
	header := make(http.Header)
	header.Add("Origin", "https://open.spotify.com")

	oauthToken, err := getOAuthToken()
	if err != nil {
		return nil, err
	}

	csrfToken, err := getCSRFToken(host, port, header)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Add("oauth", oauthToken)
	params.Add("csrf", csrfToken)

	w := &Webhelper{
		host:   host,
		port:   port,
		params: params,
		header: header,
	}
	return w, nil
}

func (w *Webhelper) Status() (*StatusResp, error) {
	params := url.Values{}
	params.Add("returnafter", "1")
	params.Add("returnon", "login,logout,play,pause,error,ap")
	resp, err := w.requestJSON("/remote/status.json", params)
	if err != nil {
		return nil, err
	}

	var data StatusResp
	if err := json.Unmarshal(resp, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

func (w *Webhelper) Play(spotifyURI string) (*StatusResp, error) {
	params := url.Values{}
	params.Add("uri", spotifyURI)
	params.Add("context", spotifyURI)
	resp, err := w.requestJSON("/remote/play.json", params)
	if err != nil {
		return nil, err
	}

	var data StatusResp
	if err := json.Unmarshal(resp, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

func (w *Webhelper) Pause() (*StatusResp, error) {
	params := url.Values{}
	params.Add("pause", "true")
	resp, err := w.requestJSON("/remote/pause.json", params)
	if err != nil {
		return nil, err
	}

	var data StatusResp
	if err := json.Unmarshal(resp, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

func (w *Webhelper) Resume() (*StatusResp, error) {
	params := url.Values{}
	params.Add("pause", "false")
	resp, err := w.requestJSON("/remote/pause.json", params)
	if err != nil {
		return nil, err
	}

	var data StatusResp
	if err := json.Unmarshal(resp, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

func (w *Webhelper) Version() (*VersionResp, error) {
	params := url.Values{}
	params.Add("service", "remote")
	resp, err := w.requestJSON("/service/version.json", params)
	if err != nil {
		return nil, err
	}

	var data VersionResp
	if err := json.Unmarshal(resp, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

func (w *Webhelper) requestJSON(path string, addParams url.Values) ([]byte, error) {
	params := url.Values{}
	for key, vals := range w.params {
		for _, val := range vals {
			params.Add(key, val)
		}
	}
	for key, vals := range addParams {
		for _, val := range vals {
			params.Add(key, val)
		}
	}
	url := generateURL(w.host, w.port, path)
	return requestJSON(url, params, w.header)
}

func getOAuthToken() (string, error) {
	resp, err := requestJSON("http://open.spotify.com/token", nil, nil)
	if err != nil {
		return "", err
	}

	var data oauthTokenResp
	if err := json.Unmarshal(resp, &data); err != nil {
		return "", err
	}
	return data.T, nil
}

func getCSRFToken(host string, port int, header http.Header) (string, error) {
	url := generateURL(host, port, "/simplecsrf/token.json")
	resp, err := requestJSON(url, nil, header)
	if err != nil {
		return "", err
	}

	var data csrfTokenResp
	if err := json.Unmarshal(resp, &data); err != nil {
		return "", err
	}
	return data.Token, nil
}

func requestJSON(url string, params url.Values, header http.Header) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header = header
	req.URL.RawQuery = params.Encode()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func generateURL(host string, port int, path string) string {
	return fmt.Sprintf("https://%s:%d%s", host, port, path)
}

func generateRandomHostname() string {
	return generateRandomString(10) + ".spotilocal.com"
}

func generateRandomString(n int) string {
	const asciiLowercase = "abcdefghijklmnopqrstuvwxyz"

	b := make([]byte, n)
	for i := range b {
		b[i] = asciiLowercase[rand.Intn(len(asciiLowercase))]
	}
	return string(b)
}

type VersionResp struct {
	Version       int    `json:"version"`
	ClientVersion string `json:"client_version"`
}

type StatusResp struct {
	VersionResp
	Playing         bool               `json:"playing"`
	Shuffle         bool               `json:"shuffle"`
	Repeat          bool               `json:"repeat"`
	PlayEnabled     bool               `json:"play_enabled"`
	PrevEnabled     bool               `json:"prev_enabled"`
	NextEnabled     bool               `json:"next_enabled"`
	Context         interface{}        `json:"context"`
	PlayingPosition float64            `json:"playing_position"`
	ServerTime      int                `json:"server_time"`
	Volume          float64            `json:"volume"`
	Online          bool               `json:"online"`
	OpenGraphState  OpenGraphStateResp `json:"open_graph_state"`
	Running         bool               `json:"running"`
	Track           TrackResp          `json:"track"`
}

type TrackResp struct {
	Length         int                `json:"length"`
	TrackType      string             `json:"track_type"`
	TrackResource  TrackResourceResp  `json:"track_resource"`
	ArtistResource ArtistResourceResp `json:"artist_resource"`
	AlbumResource  AlbumResourceResp  `json:"album_resource"`
}

type TrackResourceResp struct {
	Name     string       `json:"name"`
	URI      string       `json:"uri"`
	Location LocationResp `json:"location"`
}

type ArtistResourceResp struct {
	Name     string       `json:"name"`
	URI      string       `json:"uri"`
	Location LocationResp `json:"location"`
}

type AlbumResourceResp struct {
	Name     string       `json:"name"`
	URI      string       `json:"uri"`
	Location LocationResp `json:"location"`
}

type LocationResp struct {
	OG string `json:"og"`
}

type OpenGraphStateResp struct {
	PrivateSession  bool `json:"private_session"`
	PostingDisabled bool `json:"posting_disabled"`
}

type oauthTokenResp struct {
	T string `json:"t"`
}

type csrfTokenResp struct {
	Token string `json:"token"`
}
