package shorty

// Config primary application configuration
type Config struct {
	Address      string // listen address and port, split by colon
	WriteTimeout int    // write timeout in seconds
	ReadTimeout  int    // read timeout in seconds
	IdleTimeout  int    // idle timeout in seconds
	DatabaseFile string // path to database file (sqlite)
	SQLiteFlags  string // to be used in combination with database-file
}

// NewConfig with default values.
func NewConfig() *Config {
	return &Config{
		Address:      "127.0.0.1:8000",
		WriteTimeout: 30,
		ReadTimeout:  10,
		IdleTimeout:  60,
		DatabaseFile: "/var/lib/shorty/shorty.sqlite",
		SQLiteFlags:  "?_busy_timeout=5000&cache=shared&mode=rwc",
	}
}
