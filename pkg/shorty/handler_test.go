package shorty

import (
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
	handler = NewHandler()
	assert.NotNil(t, handler)
}

func TestHandlerSlash(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	assert.Nil(t, err)

	rr := httptest.NewRecorder()
	h := http.HandlerFunc(handler.Slash())

	h.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHandlerCreate(t *testing.T) {
	DeleteDatabaseFile(t)
	p, _ := NewPersistence(&Config{DatabaseFile: databaseFile})

	payload := strings.NewReader(fmt.Sprintf("{ \"url\": \"%s\" }", longURL))

	r := mux.NewRouter()
	r.HandleFunc("/{short}", handler.Create(p))

	ts := httptest.NewServer(r)
	res, err := http.Post(fmt.Sprintf("%s/%s", ts.URL, short), "application/json", payload)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, res.StatusCode)
}

func TestHandlerRead(t *testing.T) {
	p, _ := NewPersistence(&Config{DatabaseFile: databaseFile})

	r := mux.NewRouter()
	r.HandleFunc("/{short}", handler.Read(p))

	ts := httptest.NewServer(r)
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/%s", ts.URL, short), nil)

	transport := http.Transport{}
	res, err := transport.RoundTrip(req)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusTemporaryRedirect, res.StatusCode)
	assert.Equal(t, longURL, res.Header.Get("location"))
}
