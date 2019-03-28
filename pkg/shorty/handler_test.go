package shorty

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/stretchr/testify/assert"
)

var handler *Handler

func TestHandlerMapErr(t *testing.T) {
	err := fmt.Errorf("error")
	assert.Equal(t, gin.H{"err": err, "msg": err.Error()}, handler.mapErr(err))
}

func TestHandlerValidateURL(t *testing.T) {
	req, err := http.NewRequest("GET", "http://shorty.com/shorty", nil)
	assert.Nil(t, err)

	assert.Nil(t, handler.validateURL(req, "http://long.url.com/path"))
	assert.NotNil(t, handler.validateURL(req, ""))
	assert.NotNil(t, handler.validateURL(req, "http://127.0.0.1/path"))
	assert.NotNil(t, handler.validateURL(req, "http://localhost/path"))
	assert.NotNil(t, handler.validateURL(req, "http://shorty.com/shorty"))
}

func TestHandlerNew(t *testing.T) {
	DeleteDatabaseFile(t)
	p, _ := NewPersistence(&Config{DatabaseFile: databaseFile})

	handler = NewHandler(p)
	assert.NotNil(t, handler)
}

func recorderServeHTTP(router *gin.Engine, req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

func TestHandlerSlash(t *testing.T) {
	router := gin.Default()
	router.GET("/", handler.Slash)

	req, err := http.NewRequest("GET", "/", nil)
	assert.Nil(t, err)

	rr := recorderServeHTTP(router, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "{\"app\":\"shorty\"}", rr.Body.String())
}

func TestHandlerCreateWithoutShort(t *testing.T) {
	router := gin.Default()
	router.POST("/", handler.Create)

	payload := strings.NewReader(fmt.Sprintf("{\"url\":\"%s\"}", longURL))
	req, err := http.NewRequest("POST", "/", payload)
	assert.Nil(t, err)

	rr := recorderServeHTTP(router, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandlerCreateWithInvalidBody(t *testing.T) {
	router := gin.Default()
	router.POST("/:short", handler.Create)

	payload := strings.NewReader("bogus")
	req, err := http.NewRequest("POST", fmt.Sprintf("/%s", short), payload)
	assert.Nil(t, err)

	rr := recorderServeHTTP(router, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
}

func TestHandlerCreate(t *testing.T) {
	router := gin.Default()
	router.POST("/:short", handler.Create)

	payload := strings.NewReader(fmt.Sprintf("{\"url\":\"%s\"}", longURL))
	req, err := http.NewRequest("POST", fmt.Sprintf("/%s", short), payload)
	assert.Nil(t, err)

	rr := recorderServeHTTP(router, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
}

func TestHandlerCreateAlreadyExists(t *testing.T) {
	router := gin.Default()
	router.POST("/:short", handler.Create)

	payload := strings.NewReader(fmt.Sprintf("{\"url\":\"%s\"}", longURL))
	req, err := http.NewRequest("POST", fmt.Sprintf("/%s", short), payload)
	assert.Nil(t, err)

	rr := recorderServeHTTP(router, req)
	assert.Equal(t, http.StatusConflict, rr.Code)
}

func TestHandlerReadWithoutShort(t *testing.T) {
	router := gin.Default()
	router.GET("/", handler.Read)

	req, err := http.NewRequest("GET", "/", nil)
	assert.Nil(t, err)

	rr := recorderServeHTTP(router, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandlerReadNotFound(t *testing.T) {
	router := gin.Default()
	router.GET("/:short", handler.Read)

	req, err := http.NewRequest("GET", "/notfound", nil)
	assert.Nil(t, err)

	rr := recorderServeHTTP(router, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestHandlerRead(t *testing.T) {
	router := gin.Default()
	router.GET("/:short", handler.Read)

	req, err := http.NewRequest("GET", fmt.Sprintf("/%s", short), nil)
	assert.Nil(t, err)

	rr := recorderServeHTTP(router, req)

	assert.Equal(t, http.StatusTemporaryRedirect, rr.Code)
	assert.Equal(t, longURL, rr.Result().Header.Get("location"))
}
