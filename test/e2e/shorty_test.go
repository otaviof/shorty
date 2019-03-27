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

// TestShorty drives the tests against the app.
func TestShorty(t *testing.T) {
	t.Run("START", start)
	t.Run("GET on application root", getSlash)
	t.Run("GET on metrics endpoint", getMetrics)
	t.Run("POST URL using short string as sub-path", postShort)
	t.Run("REDIRECT after GET on short string sub-path", getShort)
	t.Run("STOP", stop)
}

// start application and wait for start-up.
func start(t *testing.T) {
	var err error

	_ = os.Remove(config.DatabaseFile)
	app, err = shorty.NewShorty(config)
	assert.Nil(t, err)

	go app.Run()
	waitForStart(t)
}

// testURL returns the test url, using global config.
func testURL() string {
	return fmt.Sprintf("http://%s", config.Address)
}

// waitForStart tries to GET the root of the app until successful or timeout.
func waitForStart(t *testing.T) {
	intervalSleep := 1 // sleep between attempts
	timeout := 15      // total time to wait for

	ready := make(chan bool, 1)
	defer close(ready)

	go func(t *testing.T) {
		for {
			req, _ := http.NewRequest("GET", testURL(), nil)
			transport := http.Transport{}
			res, err := transport.RoundTrip(req)

			if err == nil && http.StatusOK == res.StatusCode {
				ready <- true
				return
			}

			t.Logf("# Failed reaching Shorty, retry in '%d' second.", intervalSleep)
			time.Sleep(time.Duration(intervalSleep) * time.Second)
		}
	}(t)

	t.Logf("Shorty start-up timeout '%d' seconds", timeout)
	timer := time.NewTimer(time.Duration(timeout) * time.Second)
	defer timer.Stop()

	select {
	case <-ready:
		t.Logf("Shorty is ready to receive requests!")
	case <-timer.C:
		t.Fatal("Timeout on waiting for Shorty!")
	}
}

// stop wrapper for shutdown call.
func stop(t *testing.T) {
	app.Shutdown()
}

// readBody extract the bytes from response body.
func readBody(t *testing.T, body io.ReadCloser) []byte {
	bodyBytes, err := ioutil.ReadAll(body)
	assert.Nil(t, err)
	err = body.Close()
	assert.Nil(t, err)

	return bodyBytes
}

// marshalShortened wrapper for marshaling body-bytes in a Shortened object.
func marshalShortened(t *testing.T, body []byte) *shorty.Shortened {
	var shortened shorty.Shortened
	err := json.Unmarshal(body, &shortened)
	assert.Nil(t, err)
	return &shortened
}

// roundTrip executes a request.
func roundTrip(t *testing.T, req *http.Request) *http.Response {
	transport := http.Transport{}
	res, err := transport.RoundTrip(req)

	t.Logf("%s request on '%s' returned code '%d'", req.Method, req.URL.String(), res.StatusCode)
	assert.Nil(t, err)

	return res
}

// getSlash primary app endpoint.
func getSlash(t *testing.T) {
	req, err := http.NewRequest("GET", testURL(), nil)
	assert.Nil(t, err)

	res := roundTrip(t, req)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "{\"app\":\"shorty\"}", string(readBody(t, res.Body)))
}

// getMetrics prometheus compatible endpoint.
func getMetrics(t *testing.T) {
	metricsURL := fmt.Sprintf("%s/metrics", testURL())
	req, err := http.NewRequest("GET", metricsURL, nil)
	assert.Nil(t, err)

	res := roundTrip(t, req)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.True(t, len(readBody(t, res.Body)) > 0)

}

// postShort drives the tests for the post actions.
func postShort(t *testing.T) {
	postURL := fmt.Sprintf("%s/%s", testURL(), shortURL)

	postShortEmptyBody(t, postURL)
	postShortIncompleteBody(t, postURL)
	postShortCompleteBody(t, postURL)
}

// postShortEmptyBody post a empty (nil) body.
func postShortEmptyBody(t *testing.T, postURL string) {
	req, err := http.NewRequest("POST", postURL, nil)
	assert.Nil(t, err)

	t.Logf("Posting with empty JSON body (%s)", req.URL.String())
	res := roundTrip(t, req)
	t.Logf("POST returned body: '%s'", readBody(t, res.Body))

	assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
}

// postShortIncompleteBody post with a valid json, but not related to expected contract.
func postShortIncompleteBody(t *testing.T, postURL string) {
	body := []byte(fmt.Sprintf("{\"other\":\"%s\"}", longURL))
	req, err := http.NewRequest("POST", postURL, bytes.NewBuffer(body))
	assert.Nil(t, err)

	t.Logf("Posting with incomplete JSON body (%s)", req.URL.String())
	res := roundTrip(t, req)
	t.Logf("POST returned body: '%s'", readBody(t, res.Body))

	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

// postShortCompleteBody executes a post request with expected payload.
func postShortCompleteBody(t *testing.T, postURL string) {
	body := []byte(fmt.Sprintf("{\"url\":\"%s\"}", longURL))
	req, err := http.NewRequest("POST", postURL, bytes.NewBuffer(body))
	assert.Nil(t, err)

	t.Logf("Posting with complete/correct JSON body (%s)", req.URL.String())
	res := roundTrip(t, req)

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

// getShort drives the tests for get related actions.
func getShort(t *testing.T) {
	getShortNoContent(t)
	getShortExisting(t)
}

// getShortNoContent tries to get a non-existing short string.
func getShortNoContent(t *testing.T) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/bogus", testURL()), nil)
	assert.Nil(t, err)

	t.Logf("Get on bogus URL (%s)", req.URL.String())
	res := roundTrip(t, req)

	assert.Equal(t, http.StatusNoContent, res.StatusCode)
}

// getShortExisting retrieve a existing shortened object.
func getShortExisting(t *testing.T) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", testURL(), shortURL), nil)
	assert.Nil(t, err)

	t.Logf("Get existing short sub-path (%s)", req.URL.String())
	res := roundTrip(t, req)

	t.Log("Asserting redirect to original URL")
	assert.Equal(t, http.StatusTemporaryRedirect, res.StatusCode)
	assert.Equal(t, longURL, res.Header.Get("location"))
}
