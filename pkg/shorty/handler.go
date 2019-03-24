package shorty

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
)

// Handler http endpoint handlers.
type Handler struct {
	persistence *Persistence // persistence instance
}

// Slash or root, just shows the app name.
func (h *Handler) Slash() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		h.writeJSON(w, map[string]string{"app": "shorty"})
	})
}

// Create a new Shortened and store in database.
func (h *Handler) Create() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var shortened *Shortened
		var short string
		var err error

		w.Header().Set("content-type", "application/json; charset=UTF-8")

		if short = h.extractShort(r); short == "" {
			w.WriteHeader(http.StatusBadRequest)
			h.encodeErr(w, fmt.Errorf("short string is not found as sub-path"))
			return
		}
		if shortened, err = h.extractShortened(w, r); err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			h.encodeErr(w, err)
			return
		}
		if err = h.validateURL(r, shortened.URL); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			h.encodeErr(w, err)
			return
		}

		shortened.Short = short
		shortened.CreatedAt = time.Now().Unix()

		log.Printf("Saving short string '%s' for URL '%s'", shortened.Short, shortened.URL)
		if err = h.persistence.Write(shortened); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.encodeErr(w, err)
			return
		}

		log.Printf("Successfully stored URL!")
		w.WriteHeader(http.StatusCreated)
		h.writeJSON(w, shortened)
	})
}

// Read long URL from database, based in short string, and execute the redirect.
func (h *Handler) Read() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var short string
		var shortened *Shortened
		var err error

		if short = h.extractShort(r); short == "" {
			w.WriteHeader(http.StatusBadRequest)
			h.encodeErr(w, fmt.Errorf("short string is not found as sub-path"))
			return
		}

		log.Printf("Searching for long URL for short string '%s'", short)
		if shortened, err = h.retrieve(w, r, short); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.encodeErr(w, err)
			return
		}
		if shortened == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		log.Printf("Short string '%s' redirects to URL '%s'", short, shortened.URL)
		w.Header().Set("location", shortened.URL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	})
}

// encodeErr based in informed error it encodes the data to json and write it back as answer. To be
// used on runtime errors.
func (h *Handler) encodeErr(w http.ResponseWriter, err error) {
	log.Printf("[ERROR] %s", err)
	errData := map[string]interface{}{"err": err, "msg": err.Error()}

	if err := json.NewEncoder(w).Encode(errData); err != nil {
		log.Fatalf("encoding error response: '%#v'", err)
	}
}

// extractVar extract a given variable from mux.Vars, like for instance sub-path parts.
func (h *Handler) extractVar(r *http.Request, name string) (string, error) {
	reqVars := mux.Vars(r)
	if _, found := reqVars[name]; !found {
		return "", fmt.Errorf("variable named '%s' not found in request", name)
	}
	return reqVars[name], nil
}

// extractShort short string used in sub-path.
func (h *Handler) extractShort(r *http.Request) string {
	var short string
	var err error

	if short, err = h.extractVar(r, "short"); err != nil {
		log.Printf("Error on extracting 'short': '%s'", err)
		return ""
	}

	return short
}

// readBody bytes, dealing with ioutil.
func (h *Handler) readBody(body io.ReadCloser) ([]byte, error) {
	var bodyBytes []byte
	var err error

	if bodyBytes, err = ioutil.ReadAll(body); err != nil {
		return nil, err
	}
	if err = body.Close(); err != nil {
		return nil, err
	}

	return bodyBytes, nil
}

// extractShortened instance from request body.
func (h *Handler) extractShortened(w http.ResponseWriter, r *http.Request) (*Shortened, error) {
	var shortened *Shortened
	var bodyBytes []byte
	var err error

	if bodyBytes, err = h.readBody(r.Body); err != nil {
		log.Printf("Error on reading body bytes: '%s'", err)
		return nil, err
	}
	if err = json.Unmarshal(bodyBytes, &shortened); err != nil {
		log.Printf("Error on marshaling bytes to Shortened instance: '%s'", err)
		return nil, err
	}

	return shortened, nil
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

// retrieve shortned instance based on short string, from database.
func (h *Handler) retrieve(w http.ResponseWriter, r *http.Request, short string) (*Shortened, error) {
	var shortened *Shortened
	var err error

	if shortened, err = h.persistence.Read(short); err != nil {
		if h.persistence.IsErrNoRows(err) {
			log.Printf("No URL found for '%s' short string", short)
			return nil, nil
		}

		log.Printf("Error on reading persisted data: '%s'", err)
		w.WriteHeader(http.StatusInternalServerError)
		h.encodeErr(w, err)
		return nil, err
	}

	return shortened, nil
}

// writeJSON using response-writer instance to marshal any payload.
func (h *Handler) writeJSON(w http.ResponseWriter, payload interface{}) {
	var answer []byte
	var err error

	if answer, err = json.Marshal(payload); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.encodeErr(w, err)
	}
	w.Write(answer)
}

// NewHandler creates a new handler instance.
func NewHandler(persistence *Persistence) *Handler {
	return &Handler{persistence: persistence}
}
