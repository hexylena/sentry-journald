package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/urfave/cli/v2"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Get("/", indexPage)

	r.Get("/test", testPage)

	r.Get("/issues/{msg}", issuePage)

	r.Post("/api/{projectID}/envelope/", func(w http.ResponseWriter, r *http.Request) {
		processSentryRequest(w, r)
	})

	app := &cli.App{
		Name:  "sentry-journald",
		Usage: "Log your sentry errors directly into your journald",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "port",
				Value:   "8008",
				Usage:   "port to listen on",
				EnvVars: []string{"PORT"},
			},
		},
		Action: func(cctx *cli.Context) error {
			hostname, _ := os.Hostname()
			fmt.Printf("Configure your sentry project to use this server as the DSN endpoint\n\n")
			fmt.Printf("http://my-project-name@%s:%s/1\n\n", hostname, cctx.String("port"))
			fmt.Printf("Note that the public key (first component) may be set to any string, we recommend using it as a project name. The project ID (the numeric trailing component) may be set to any number to disambiguate projects, as there is no built-in database that would use the project ID.\n")
			http.ListenAndServe(":"+cctx.String("port"), r)
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}
