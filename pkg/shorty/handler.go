package shorty

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler http endpoint handlers.
type Handler struct {
	persistence *Persistence // persistence instance
}

// Slash or root, just shows the app name.
func (h *Handler) Slash(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"app": "shorty"})
}

// Create a new Shortened and store in database.
func (h *Handler) Create(c *gin.Context) {
	var shortened Shortened
	var short string
	var err error

	if short = c.Param("short"); short == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, fmt.Errorf("short is not found as sub-path"))
		return
	}

	if err = c.BindJSON(&shortened); err != nil {
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, err)
		return
	}

	if err = h.validateURL(c.Request, shortened.URL); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	shortened.Short = short
	shortened.CreatedAt = time.Now().Unix()

	log.Printf("Saving short string '%s' for URL '%s'", shortened.Short, shortened.URL)
	if err = h.persistence.Write(c.Request.Context(), &shortened); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	log.Printf("Successfully stored URL!")
	c.JSONP(http.StatusCreated, shortened)
}

// Read long URL from database, based in short string, and execute the redirect.
func (h *Handler) Read(c *gin.Context) {
	var short string
	var shortened *Shortened
	var err error

	if short = c.Param("short"); short == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, fmt.Errorf("short is not found as sub-path"))
		return
	}

	log.Printf("Searching for long URL for short string '%s'", short)
	if shortened, err = h.persistence.Read(c.Request.Context(), short); err != nil && !h.persistence.IsErrNoRows(err) {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	if shortened == nil {
		log.Printf("No shortened URL is found for '%s' short string", short)
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	log.Printf("Short string '%s' redirects to URL '%s'", short, shortened.URL)
	c.Header("location", shortened.URL)
	c.JSONP(http.StatusTemporaryRedirect, shortened)
}

// validateURL check if informed URL is valid and does not point to the same redirect service.
func (h *Handler) validateURL(r *http.Request, longURL string) error {
	var parsed *url.URL
	var err error

	if longURL == "" {
		return fmt.Errorf("empty URL informed")
	}
	if parsed, err = url.Parse(longURL); err != nil {
		return err
	}

	hostname := parsed.Hostname()
	if hostname == r.Host {
		return fmt.Errorf("redirects to the same service hostname ('%s') is not allowed", hostname)
	}
	if hostname == "127.0.0.1" || hostname == "localhost" {
		return fmt.Errorf("redirects to localhost are not allowed")
	}

	return nil
}

// NewHandler creates a new handler instance.
func NewHandler(persistence *Persistence) *Handler {
	return &Handler{persistence: persistence}
}
