package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/ssgreg/journald"
	"github.com/urfave/cli/v2"
)

type SentryHeaderSdk struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type SentryHeaderTrace struct {
	Environment string `json:"environment"`
	Release     string `json:"release"`
	Public_key  string `json:"public_key"`
	Trace_id    string `json:"trace_id"`
}
type SentryHeader struct {
	Event_id string            `json:"event_id"`
	Sent_at  string            `json:"sent_at"`
	Sdk      SentryHeaderSdk   `json:"sdk"`
	Trace    SentryHeaderTrace `json:"trace"`
}

type SentryEnvelope struct {
	Type string `json:"type"`
}

type SentryBreadcrumb struct {
	Category  string `json:"category"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
}

type SentryBreadcrumbContainer struct {
	Values []SentryBreadcrumb `json:"values"`
}

type SentryRequest struct {
	Url     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}

type SentryStackTraceFrame struct {
	Filename string `json:"filename"`
	Function string `json:"function"`
	InApp    bool   `json:"in_app"`
	Lineno   int    `json:"lineno"`
	Colno    int    `json:"colno"`
}

type SentryStackTrace struct {
	Frames []SentryStackTraceFrame `json:"frames"`
}

type SentryExceptionItem struct {
	Type       string           `json:"type"`
	Value      string           `json:"value"`
	Stacktrace SentryStackTrace `json:"stacktrace"`
}

type SentryException struct {
	Values []SentryExceptionItem `json:"values"`
}

type SentryEvent struct {
	Message    string                 `json:"message"`
	Level      string                 `json:"level"`
	Event_id   string                 `json:"event_id"`
	Timestamp  interface{}            `json:"timestamp"`
	Contexts   map[string]interface{} `json:"contexts"`
	Exception  SentryException        `json:"exception"`
	Stacktrace SentryStackTrace       `json:"stacktrace"`
	// Breadcrumbs  []SentryBreadcrumb `json:"breadcrumbs"`
	Modules     map[string]string `json:"modules"`
	Release     string            `json:"release"`
	Environment string            `json:"environment"`
	Server_name string            `json:"server_name"`
	Platform    string            `json:"platform"`
	Request     SentryRequest     `json:"request"`
}

func processSentryRequest(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	// These can be provided here,
	sentry_key := r.URL.Query().Get("sentry_key")
	sentry_version := r.URL.Query().Get("sentry_version")
	sentry_client := r.URL.Query().Get("sentry_client")
	// or in a header X-Sentry-Auth
	// X-Sentry-Auth: Sentry sentry_key=gtn-py, sentry_version=7, sentry_client=sentry.python/1.40.6
	x_sentry_auth := r.Header.Get("X-Sentry-Auth")
	if x_sentry_auth != "" {
		// parse it
		// X-Sentry-Auth: Sentry sentry_key=gtn-py, sentry_version=7, sentry_client=sentry.python/1.40.6
		//strip 'Sentry '
		if string(x_sentry_auth[0:7]) == "Sentry " {
			x_sentry_auth = x_sentry_auth[7:]
		}

		parts := strings.Split(x_sentry_auth, ", ")
		for _, part := range parts {
			kv := strings.Split(part, "=")
			switch kv[0] {
			case "sentry_key":
				sentry_key = kv[1]
			case "sentry_version":
				sentry_version = kv[1]
			case "sentry_client":
				sentry_client = kv[1]
			}
		}
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request", http.StatusInternalServerError)
	}

	// Split into 3 on newlines
	// 1. Header
	// 2. Event
	// 3. Context

	// it might be gzip compressed
	// check the header
	// if it's gzip, decompress it
	content_encoding := r.Header.Get("Content-Encoding")
	if content_encoding == "gzip" {
		// decompress
		reader, err := gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			http.Error(w, "Error decompressing request", http.StatusBadRequest)
		}
		defer reader.Close()
		body, err = ioutil.ReadAll(reader)
		if err != nil {
			http.Error(w, "Error decompressing request", http.StatusBadRequest)
		}
	}
	v := bytes.Split(body, []byte("\n"))

	header := SentryHeader{}
	err = json.Unmarshal(v[0], &header)
	if err != nil {
		http.Error(w, "Error parsing header", http.StatusBadRequest)
	}

	msg_type := SentryEnvelope{}
	err = json.Unmarshal(v[1], &msg_type)
	if err != nil {
		http.Error(w, "Error parsing type", http.StatusBadRequest)
	}

	// Explicitly choosing to discard session and transaction information.
	// User reports and events are kept.
	if msg_type.Type == "session" || msg_type.Type == "transaction" {
		w.Write([]byte("{}"))
		return
	}

	event2 := SentryEvent{}
	err = json.Unmarshal(v[2], &event2)
	if err != nil {
		http.Error(w, "Error parsing event", http.StatusBadRequest)
	}

	journal_metadata := map[string]interface{}{
		"SENTRY_KEY":          sentry_key,
		"SENTRY_VERSION":      sentry_version,
		"SENTRY_CLIENT":       sentry_client,
		"SYSLOG_IDENTIFIER":   "sentry",
		"PROJECT_ID":          projectID,
		"MESSAGE_ID":          header.Event_id,
		"SENTRY_ENVIRONMENT":  event2.Environment,
		"SENTRY_RELEASE":      event2.Release,
		"SENTRY_DIST":         event2.Release,
		"SENTRY_PLATFORM":     event2.Platform,
		"SENTRY_TIMESTAMP":    event2.Timestamp,
		"SENTRY_SERVER_NAME":  event2.Server_name,
		"REQUEST_URL":         event2.Request.Url,
		"REQUEST_REMOTE_ADDR": r.RemoteAddr,
		"REMOTE_ADDR":         r.RemoteAddr,
		"REQUEST_METHOD":      r.Method,
	}

	// process whatever we have.
	journal_metadata["REQUEST_HEADERS"], _ = json.Marshal(event2.Request.Headers)
	journal_metadata["SENTRY_CONTEXTS"], _ = json.Marshal(event2.Contexts)

	var message string
	if event2.Message != "" {
		message = event2.Message
	} else {
		if len(event2.Exception.Values) > 0 {
			// Will it ever be more??
			value := event2.Exception.Values[0]
			message = fmt.Sprintf("%s: %s", value.Type, value.Value)

			// Prepend the stacktrace
			for _, frame := range value.Stacktrace.Frames {
				message = fmt.Sprintf("[%s:%d:%d] %s",
					frame.Filename, frame.Lineno, frame.Colno, message)
			}
		}
	}

	var log_message string
	log_message = fmt.Sprintf("[%s] (proj=%s env=%s) %s", msg_type.Type, sentry_key, event2.Environment, message)
	// if there's a release, add it to the message

	var log_level journald.Priority
	switch event2.Level {
	case "debug":
		log_level = journald.PriorityDebug
	case "info":
		log_level = journald.PriorityInfo
	case "warning":
		log_level = journald.PriorityWarning
	case "error":
		log_level = journald.PriorityErr
	case "fatal":
		log_level = journald.PriorityCrit
	default:
		log_level = journald.PriorityNotice
	}

	journald.Send(log_message, log_level, journal_metadata)

	if header.Event_id != "" {
		w.Write([]byte(fmt.Sprintf("{\"id\": \"%s\"}", header.Event_id)))
	} else {
		w.Write([]byte("{}"))
	}
}

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

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	r.Post("/api/{projectID}/envelope/", func(w http.ResponseWriter, r *http.Request) {
		processSentryRequest(w, r)
	})

	app := &cli.App{
		Name:  "sentry-journald",
		Usage: "Log your sentry errors directly into your journald",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "port",
				Value: "8000",
				Usage: "port to listen on",
			},
		},
		Action: func(cctx *cli.Context) error {
			fmt.Printf("Serving on :%s\n", cctx.String("port"))
			hostname, _ := os.Hostname()
			fmt.Printf("Configure your sentry project to use this server as the DSN endpoint\n\n")
			fmt.Printf("http://my-project-name@%s:%s/1\n\n", hostname, cctx.String("port"))
			fmt.Printf("Note that the public key (first component) may be set to any string, we recommend using it as a project name. The project ID (the numeric trailing component) may be set to any number to disambiguate projects, as there is no built-in database that would use the project ID.")
			http.ListenAndServe(":"+cctx.String("port"), r)
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}
