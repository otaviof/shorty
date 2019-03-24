package shorty

// Shortened represents a entry in persistence store.
type Shortened struct {
	Short     string `json:"short,omitempty"`      // short URL
	URL       string `json:"url"`                  // original URL
	CreatedAt int64  `json:"created_at,omitempty"` // created timestamp
}
