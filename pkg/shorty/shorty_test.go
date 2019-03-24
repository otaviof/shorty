package shorty

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var shorty *Shorty

func TestShortyNewShorty(t *testing.T) {
	var err error

	config := NewConfig()
	config.DatabaseFile = "../../.data/shorty.sqlite"
	shorty, err = NewShorty(config)

	assert.Nil(t, err)
	assert.NotNil(t, shorty)
}

func TestShortyRun(t *testing.T) {
	t.Log("Running Shorty in background")
	go shorty.Run()

	t.Log("Waiting a few seconds for app bootstrap...")
	time.Sleep(5 * time.Second)

	config := NewConfig()
	res, err := http.Get(fmt.Sprintf("http://%s/", config.Address))

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	t.Log("Shuting down app")
	shorty.Shutdown()
}
