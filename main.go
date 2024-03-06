package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	// "strings"
	// "os"
	// "runtime"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/ssgreg/journald"
)

func fetch(key string, m map[string]interface{}) string {
	if val, ok := m[key]; ok {
		return val.(string)
	}
	return "nil"
}

// {"event_id":"5d7e101599854abd8e8e8b85a3f07813","sent_at":"2024-03-06T11:33:58.300Z","sdk":{"name":"sentry.javascript.browser","version":"7.105.0"},"trace":{"environment":"production","release":"my-project-name@2.3.12","public_key":"password","trace_id":"3deeddbb8ba04ff4bc5f4e8d7fefdd40"}}
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
	Category string `json:"category"`
	Level    string `json:"level"`
	Message  string `json:"message"`
	Type     string `json:"type"`
	Timestamp string `json:"timestamp"`
}

type SentryBreadcrumbContainer struct {
	Values []SentryBreadcrumb `json:"values"`
}

type SentryRequest struct {
	Url string `json:"url"`
	Headers map[string]string `json:"headers"`
}

// stacktrace":{"frames":[{"filename":"http://localhost:4001/test.html","function":"onclick","in_app":true,"lineno":1,"colno":1},{"filename":"http://localhost:4001/test.html","function":"doesNotExist","in_app":true,"lineno":38,"colno":5}]}

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
	Type   string `json:"type"`
	Value  string `json:"value"`
	Stacktrace SentryStackTrace `json:"stacktrace"`
}

type SentryException struct {
	Values []SentryExceptionItem `json:"values"`
}

type SentryEvent struct {
	Message string `json:"message"`
	Level   string `json:"level"`
	Event_id string `json:"event_id"`
	Timestamp    interface{} `json:"timestamp"`
	Contexts     map[string]interface{} `json:"contexts"`
	// Stacktrace   SentryStackTrace `json:"stacktrace"`
	Exception SentryException `json:"exception"`
	Stacktrace   SentryStackTrace `json:"stacktrace"`
	// Breadcrumbs  []SentryBreadcrumb `json:"breadcrumbs"`
	Modules      map[string]string `json:"modules"`
	Release      string `json:"release"`
	Environment  string `json:"environment"`
	Server_name  string `json:"server_name"`
	Platform     string `json:"platform"`
	Request      SentryRequest `json:"request"`
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
		// https://sentry.galaxyproject.org/api/10/envelope/?sentry_key=45e0ec6e4373462b92969505df37cf40&sentry_version=7&sentry_client=sentry.javascript.browser%2F7.52.1
		projectID := chi.URLParam(r, "projectID")
		// don't super need sentry_version, sentry_client?
		sentry_key := r.URL.Query().Get("sentry_key")
		sentry_version := r.URL.Query().Get("sentry_version")
		sentry_client := r.URL.Query().Get("sentry_client")

		// Get post data
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
			// fmt.Printf("body=%s\n", body)
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

		var event map[string]interface{}
		err = json.Unmarshal(v[2], &event)
		if err != nil {
			http.Error(w, "Error parsing body", http.StatusBadRequest)
		}
		// fmt.Printf("event=%s\n", v[2])

		event2 := SentryEvent{}
		err = json.Unmarshal(v[2], &event2)
		if err != nil {
			http.Error(w, "Error parsing event", http.StatusBadRequest)
		}
		// fmt.Printf("event2=%s\n", event2.Exception)
		// fmt.Println(header, msg_type, event)

		journal_metadata := map[string]interface{}{
			"SENTRY_KEY":        sentry_key,
			"SENTRY_VERSION":    sentry_version,
			"SENTRY_CLIENT":     sentry_client,
			"SYSLOG_IDENTIFIER": "sentry",
			"PROJECT_ID":        projectID,
		}

		journal_metadata["MESSAGE_ID"] = header.Event_id

		journal_metadata["SENTRY_ENVIRONMENT"] = event2.Environment
		journal_metadata["SENTRY_RELEASE"] = event2.Release
		journal_metadata["SENTRY_DIST"] = event2.Release
		journal_metadata["SENTRY_PLATFORM"] = event2.Platform
		journal_metadata["SENTRY_TIMESTAMP"] = event2.Timestamp
		journal_metadata["SENTRY_SERVER_NAME"] = event2.Server_name
		journal_metadata["REQUEST_URL"] = event2.Request.Url
		journal_metadata["REQUEST_HEADERS"], _ = json.Marshal(event2.Request.Headers)
		// ip address
		journal_metadata["REQUEST_REMOTE_ADDR"] = r.RemoteAddr
		journal_metadata["REMOTE_ADDR"] = r.RemoteAddr
		journal_metadata["REQUEST_METHOD"] = r.Method

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
		log_message = fmt.Sprintf("[%s] (proj=%s env=%s) %s", msg_type.Type, fetch("project", event), fetch("environment", event), message)
		// if there's a release, add it to the message

		if _, ok := event["level"]; ok {
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
		}

		if header.Event_id != "" {
			w.Write([]byte(fmt.Sprintf("{\"id\": \"%s\"}", header.Event_id)))
		} else {
			w.Write([]byte("{}"))
		}

	})

	fmt.Println("Serving on :8000")
	http.ListenAndServe(":8000", r)
}
