package shorty

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"

	"github.com/stretchr/testify/assert"
)

var handler *Handler

func TestHandlerNew(t *testing.T) {
	DeleteDatabaseFile(t)
	p, _ := NewPersistence(&Config{DatabaseFile: databaseFile})

	handler = NewHandler(p)
	assert.NotNil(t, handler)
}

func TestHandlerEncodeErr(t *testing.T) {
	err := fmt.Errorf("error")
	rr := httptest.NewRecorder()

	handler.encodeErr(rr, err)

	assert.Equal(t, "{\"err\":{},\"msg\":\"error\"}\n", rr.Body.String())
}

func TestHandlerReadBody(t *testing.T) {
	body := []byte("string")
	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(body))
	assert.Nil(t, err)

	b, err := handler.readBody(req.Body)

	assert.Nil(t, err)
	assert.Equal(t, body, b)
}

func TestHandlerExtractShortened(t *testing.T) {
	body := []byte(fmt.Sprintf("{\"url\":\"%s\"}", longURL))
	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(body))
	assert.Nil(t, err)

	rr := httptest.NewRecorder()
	shortened, err := handler.extractShortened(rr, req)

	assert.Nil(t, err)
	assert.Equal(t, longURL, shortened.URL)
}

func TestHandlerWriteJSON(t *testing.T) {
	rr := httptest.NewRecorder()

	handler.writeJSON(rr, map[string]string{"test": "test"})

	assert.Equal(t, "{\"test\":\"test\"}", rr.Body.String())
}

func TestHandlerSlash(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	assert.Nil(t, err)

	rr := httptest.NewRecorder()
	h := http.HandlerFunc(handler.Slash())

	h.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "{\"app\":\"shorty\"}", rr.Body.String())
}

func TestHandlerPersist(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	assert.Nil(t, err)

	shortened := &Shortened{Short: "short", URL: "URL", CreatedAt: 0}
	rr := httptest.NewRecorder()

	err = handler.persist(rr, req, shortened)

	assert.Nil(t, err)
}

func TestHandlerRetrieve(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	assert.Nil(t, err)

	rr := httptest.NewRecorder()
	shortened, err := handler.retrieve(rr, req, "short")

	assert.Nil(t, err)
	assert.Equal(t, "URL", shortened.URL)
}

func TestHandlerCreate(t *testing.T) {
	payload := strings.NewReader(fmt.Sprintf("{\"url\":\"%s\"}", longURL))

	r := mux.NewRouter()
	r.HandleFunc("/{short}", handler.Create())

	ts := httptest.NewServer(r)
	res, err := http.Post(fmt.Sprintf("%s/%s", ts.URL, short), "application/json", payload)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, res.StatusCode)
}

func TestHandlerRead(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/{short}", handler.Read())

	ts := httptest.NewServer(r)
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/%s", ts.URL, short), nil)

	transport := http.Transport{}
	res, err := transport.RoundTrip(req)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusTemporaryRedirect, res.StatusCode)
	assert.Equal(t, longURL, res.Header.Get("location"))
}
