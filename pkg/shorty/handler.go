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
type Handler struct{}

// Slash or root, just shows the app name.
func (h *Handler) Slash() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json; charset=UTF-8")
		answer, err := json.Marshal(map[string]string{"app": "shorty"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.encodeErr(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(answer)
	})
}

// Create a new Shortened and store in database.
func (h *Handler) Create(p *Persistence) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var shortened *Shortened
		var short string
		var answer []byte
		var err error

		w.Header().Set("content-type", "application/json; charset=UTF-8")

		if short, err = h.extractVar(r, "short"); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			h.encodeErr(w, err)
			return
		}
		if shortened, err = h.readBody(r.Body); err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			h.encodeErr(w, err)
			return
		}

		shortened.Short = short
		shortened.CreatedAt = time.Now().Unix()

		log.Printf("Saving short string '%s' for URL '%s'", shortened.Short, shortened.URL)
		if err = p.Write(shortened); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.encodeErr(w, err)
			return
		}

		log.Printf("Successfully stored URL!")
		w.WriteHeader(http.StatusCreated)
		if answer, err = json.Marshal(shortened); err != nil {
			h.encodeErr(w, err)
			return
		}
		w.Write(answer)
	})

}

// Read long URL from database, based in short string, and execute the redirect.
func (h *Handler) Read(p *Persistence) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var short string
		var shortened *Shortened
		var err error

		w.Header().Set("content-type", "application/json; charset=UTF-8")

		if short, err = h.extractVar(r, "short"); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			h.encodeErr(w, err)
			return
		}

		log.Printf("Searching for long URL for short string '%s'", short)
		if shortened, err = p.Read(short); err != nil {
			if p.IsErrNoRows(err) {
				log.Printf("No URL found for '%s' short string", short)
				w.WriteHeader(http.StatusNoContent)
				return
			}

			log.Printf("Error on reading persisted data: '%s'", err)
			w.WriteHeader(http.StatusInternalServerError)
			h.encodeErr(w, err)
			return
		}

		log.Printf("Short string '%s' redirects to URL '%s'", short, shortened.URL)
		w.Header().Set("location", shortened.URL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	})
}

// readBody capture the body and threats as a Shortened object, trying to unmarshal it.
func (h *Handler) readBody(body io.ReadCloser) (*Shortened, error) {
	var bodyBytes []byte
	var err error

	if bodyBytes, err = ioutil.ReadAll(body); err != nil {
		return nil, err
	}
	if err = body.Close(); err != nil {
		return nil, err
	}

	shortened := Shortened{}
	if err = json.Unmarshal(bodyBytes, &shortened); err != nil {
		return nil, err
	}

	return &shortened, nil
}

// extractVar extract a given variable from mux.Vars, like for instance sub-path parts.
func (h *Handler) extractVar(r *http.Request, name string) (string, error) {
	reqVars := mux.Vars(r)
	if _, found := reqVars[name]; !found {
		return "", fmt.Errorf("variable named '%s' not found in request", name)
	}
	return reqVars[name], nil
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

// NewHandler creates a new handler instance.
func NewHandler() *Handler {
	return &Handler{}
}
