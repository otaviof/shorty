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
	if err = c.ShouldBindJSON(&shortened); err != nil {
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, h.mapErr(err))
		return
	}
	if err = h.validateURL(c.Request, shortened.URL); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, h.mapErr(err))
		return
	}

	shortened.Short = short
	shortened.CreatedAt = time.Now().Unix()

	log.Printf("Saving short string '%s' for URL '%s'", shortened.Short, shortened.URL)
	if err = h.persistence.Write(c.Request.Context(), &shortened); err != nil {
		status := http.StatusInternalServerError
		log.Printf("Persistence error: '%s'", err)
		if h.persistence.IsErrUniqueConstraint(err) {
			status = http.StatusConflict
		}
		c.AbortWithStatusJSON(status, h.mapErr(err))
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
	if shortened, err = h.persistence.Read(
		c.Request.Context(), short,
	); err != nil && !h.persistence.IsErrNoRows(err) {
		c.AbortWithStatusJSON(http.StatusInternalServerError, h.mapErr(err))
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

// List shows all shortened URLs as a array of entries.
func (h *Handler) List(c *gin.Context) {
	slice, err := h.persistence.List(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, h.mapErr(err))
	}
	log.Printf("Found '%d' shortened entries.", len(slice))
	c.JSONP(http.StatusOK, slice)
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

// mapErr include error message along side error codes.
func (h *Handler) mapErr(err error) gin.H {
	return gin.H{"err": err, "msg": err.Error()}
}

// NewHandler creates a new handler instance.
func NewHandler(persistence *Persistence) *Handler {
	return &Handler{persistence: persistence}
}
