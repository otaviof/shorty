package shorty

import "fmt"

// Config primary application configuration
type Config struct {
	Address      string // listen address and port, split by colon
	WriteTimeout int    // write timeout in seconds
	ReadTimeout  int    // read timeout in seconds
	IdleTimeout  int    // idle timeout in seconds
	DatabaseFile string // path to database file (sqlite)
	SQLiteFlags  string // to be used in combination with database-file
}

// Validate config contents.
func (c *Config) Validate() error {
	if c.Address == "" {
		return fmt.Errorf("address is empty: '%s'", c.Address)
	}
	if c.WriteTimeout <= 0 {
		return fmt.Errorf("invalid value for write-timeout: '%d'", c.WriteTimeout)
	}
	if c.ReadTimeout <= 0 {
		return fmt.Errorf("invalid value for read-timeout: '%d'", c.ReadTimeout)
	}
	if c.IdleTimeout <= 0 {
		return fmt.Errorf("invalid value for idle-timeout: '%d'", c.IdleTimeout)
	}
	if c.DatabaseFile == "" {
		return fmt.Errorf("database-file is not informed: '%s'", c.DatabaseFile)
	}
	return nil
}

// NewConfig with default values.
func NewConfig() *Config {
	return &Config{
		Address:      "127.0.0.1:8000",
		WriteTimeout: 30,
		ReadTimeout:  10,
		IdleTimeout:  60,
		DatabaseFile: "/var/lib/shorty/shorty.sqlite",
		SQLiteFlags:  "_busy_timeout=5000&cache=shared&mode=rwc",
	}
}
