package main

import (
	"strings"

	shorty "github.com/otaviof/shorty/pkg/shorty"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "shorty",
	Run:   runShortyApp,
	Short: "Yet another URL shortener. Written in Go with SQLite backend.",
	Long: `
#
# shorty
#

Shorty is a simple application to expose a REST interface to register and redirect URLs based on a
arbitrary short string.

For instance, to create a short link to "google.nl" as "ggl", execute:

	curl -X POST http://127.0.0.1:8000/shorty/shorty \
		-d '{ "url": "https://github.com/otaviof/shorty" }'

After registering the short link, you can use "curl" again to follow the redirect:

	curl -L http://127.0.0.1:8000/shorty/shorty

Backend persistence is done using SQLite, you can inform more options in command-line to use an
alternative database-file and connection flags. Short links can't be repeated, since it's a
constraint in Shorty's table.`,
}

// runShortyApp instantiage the application using runtime config.
func runShortyApp(cmd *cobra.Command, args []string) {
	var err error
	var app *shorty.Shorty

	config := bootstrapConfig()
	if err = config.Validate(); err != nil {
		panic(err)
	}

	if app, err = shorty.NewShorty(config); err != nil {
		panic(err)
	}

	app.Run()
}

// bootstrapConfig using viper, therefore environment variables can overwrite command-line flags.
func bootstrapConfig() *shorty.Config {
	return &shorty.Config{
		Address:      viper.GetString("address"),
		DatabaseFile: viper.GetString("database-file"),
		IdleTimeout:  viper.GetInt("idle-timeout"),
		ReadTimeout:  viper.GetInt("read-timeout"),
		WriteTimeout: viper.GetInt("write-timeout"),
		SQLiteFlags:  viper.GetString("sqlite-flags"),
	}
}

// init setup command-line arguments.
func init() {
	flags := rootCmd.PersistentFlags()

	// setting up rules for environment variables
	viper.SetEnvPrefix("shorty")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// command-line options
	flags.String("address", "127.0.0.1:8000", "Listen address")
	flags.String("database-file", "/var/lib/shorty/shorty.sqlite", "SQLite database file path")
	flags.Int("idle-timeout", 10, "HTTP connection idle-timeout in seconds")
	flags.Int("read-timeout", 5, "HTTP connection read-timeout in seconds")
	flags.Int("write-timeout", 30, "HTTP connection write-timeout in seconds")
	flags.String("sqlite-flags", "", "SQLite connection string flags")

	if err := viper.BindPFlags(flags); err != nil {
		panic(err)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
