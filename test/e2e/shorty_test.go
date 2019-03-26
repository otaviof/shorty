package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	shorty "github.com/otaviof/shorty/pkg/shorty"
)

const (
	longURL  = "http://github.com/otaviof"
	shortURL = "e2e"
)

var config = &shorty.Config{
	Address:      "127.0.0.1:8001",
	IdleTimeout:  15,
	ReadTimeout:  15,
	WriteTimeout: 30,
	DatabaseFile: "/var/tmp/shorty-e2e.sqlite",
	SQLiteFlags:  "",
}
var app *shorty.Shorty

func TestShorty(t *testing.T) {
	t.Run("START", start)
	t.Run("GET on application root", getSlash)
	t.Run("GET on metrics endpoint", getMetrics)
	t.Run("POST URL using short string as sub-path", postShort)
	t.Run("REDIRECT after GET on short string sub-path", getShort)
	t.Run("STOP", stop)
}

func start(t *testing.T) {
	var err error

	_ = os.Remove(config.DatabaseFile)
	app, err = shorty.NewShorty(config)
	assert.Nil(t, err)

	go app.Run()
	time.Sleep(5 * time.Second)
}

func stop(t *testing.T) {
	app.Shutdown()
}

func testURL() string {
	return fmt.Sprintf("http://%s", config.Address)
}

func readBody(t *testing.T, body io.ReadCloser) []byte {
	bodyBytes, err := ioutil.ReadAll(body)
	assert.Nil(t, err)
	err = body.Close()
	assert.Nil(t, err)

	return bodyBytes
}

func marshalShortened(t *testing.T, body []byte) *shorty.Shortened {
	var shortened shorty.Shortened
	err := json.Unmarshal(body, &shortened)
	assert.Nil(t, err)
	return &shortened
}

func getSlash(t *testing.T) {
	req, err := http.NewRequest("GET", testURL(), nil)

	transport := http.Transport{}
	res, err := transport.RoundTrip(req)

	t.Logf("GET Request on '%s' returned code '%d'", testURL(), res.StatusCode)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "{\"app\":\"shorty\"}", string(readBody(t, res.Body)))
}

func getMetrics(t *testing.T) {
	metricsURL := fmt.Sprintf("%s/metrics", testURL())
	req, err := http.NewRequest("GET", metricsURL, nil)

	transport := http.Transport{}
	res, err := transport.RoundTrip(req)

	t.Logf("GET Request on '%s' returned code '%d'", metricsURL, res.StatusCode)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.True(t, len(readBody(t, res.Body)) > 0)

}

func postShort(t *testing.T) {
	postURL := fmt.Sprintf("%s/%s", testURL(), shortURL)

	postShortEmptyBody(t, postURL)
	postShortIncompleteBody(t, postURL)
	postShortCompleteBody(t, postURL)
}

func postShortEmptyBody(t *testing.T, postURL string) {
	req, err := http.NewRequest("POST", postURL, nil)

	t.Logf("Posting with empty JSON body (%s)", req.URL.String())
	transport := http.Transport{}
	res, err := transport.RoundTrip(req)

	t.Logf("POST returned body: '%s'", readBody(t, res.Body))

	assert.Nil(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
}

func postShortIncompleteBody(t *testing.T, postURL string) {
	body := []byte(fmt.Sprintf("{\"other\":\"%s\"}", longURL))
	req, err := http.NewRequest("POST", postURL, bytes.NewBuffer(body))

	t.Logf("Posting with incomplete JSON body (%s)", req.URL.String())
	transport := http.Transport{}
	res, err := transport.RoundTrip(req)

	t.Logf("POST returned body: '%s'", readBody(t, res.Body))

	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func postShortCompleteBody(t *testing.T, postURL string) {
	body := []byte(fmt.Sprintf("{\"url\":\"%s\"}", longURL))
	req, err := http.NewRequest("POST", postURL, bytes.NewBuffer(body))

	t.Logf("Posting with complete/correct JSON body (%s)", req.URL.String())
	transport := http.Transport{}
	res, err := transport.RoundTrip(req)

	resBodyBytes := readBody(t, res.Body)
	shortened := marshalShortened(t, resBodyBytes)

	t.Logf("POST on '%s' returned code '%d'", req.URL, res.StatusCode)
	t.Logf("POST returned shortened: '%#v'", shortened)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, res.StatusCode)
	assert.Equal(t, shortURL, shortened.Short)
	assert.Equal(t, longURL, shortened.URL)
	assert.True(t, shortened.CreatedAt > 0)
}

func getShort(t *testing.T) {
	getShortNoContent(t)
	getShortExisting(t)
}

func getShortNoContent(t *testing.T) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/bogus", testURL()), nil)

	t.Logf("Get on bogus URL (%s)", req.URL.String())
	transport := http.Transport{}
	res, err := transport.RoundTrip(req)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusNoContent, res.StatusCode)
}

func getShortExisting(t *testing.T) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", testURL(), shortURL), nil)

	t.Logf("Get existing short sub-path (%s)", req.URL.String())
	transport := http.Transport{}
	res, err := transport.RoundTrip(req)

	t.Log("Asserting redirect to original URL")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusTemporaryRedirect, res.StatusCode)
	assert.Equal(t, longURL, res.Header.Get("location"))
}
