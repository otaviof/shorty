package shorty

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
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

		if short = h.extractShort(w, r); short == "" {
			return
		}
		if shortened, err = h.extractShortened(w, r); err != nil {
			return
		}
		shortened.Short = short
		shortened.CreatedAt = time.Now().Unix()

		log.Printf("Saving short string '%s' for URL '%s'", shortened.Short, shortened.URL)
		if err = h.persist(w, r, shortened); err != nil {
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

		if short = h.extractShort(w, r); short == "" {
			return
		}

		log.Printf("Searching for long URL for short string '%s'", short)
		if shortened, err = h.retrieve(w, r, short); err != nil {
			return
		}
		if shortened.URL == "" {
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
func (h *Handler) extractShort(w http.ResponseWriter, r *http.Request) string {
	var short string
	var err error

	if short, err = h.extractVar(r, "short"); err != nil {
		log.Printf("Error on extracting 'short': '%s'", err)
		w.WriteHeader(http.StatusBadRequest)
		h.encodeErr(w, err)
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
		w.WriteHeader(http.StatusUnprocessableEntity)
		h.encodeErr(w, err)
		return nil, err
	}

	if err = json.Unmarshal(bodyBytes, &shortened); err != nil {
		log.Printf("Error on marshaling bytes to Shortened instance: '%s'", err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		h.encodeErr(w, err)
		return nil, err
	}

	return shortened, nil
}

// retrieve shortned instance based on short string, from database.
func (h *Handler) retrieve(w http.ResponseWriter, r *http.Request, short string) (*Shortened, error) {
	var shortened *Shortened
	var err error

	if shortened, err = h.persistence.Read(short); err != nil {
		if h.persistence.IsErrNoRows(err) {
			log.Printf("No URL found for '%s' short string", short)
			return shortened, nil
		}

		log.Printf("Error on reading persisted data: '%s'", err)
		w.WriteHeader(http.StatusInternalServerError)
		h.encodeErr(w, err)
		return nil, err
	}

	return shortened, nil
}

// persist shortened instance to database.
func (h *Handler) persist(w http.ResponseWriter, r *http.Request, shortened *Shortened) error {
	if err := h.persistence.Write(shortened); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.encodeErr(w, err)
		return err
	}
	return nil
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
