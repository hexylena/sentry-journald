package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/ssgreg/journald"
)

// TODO: make toggleable.
var DEBUG bool

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
	if msg_type.Type == "session" || msg_type.Type == "transaction" || msg_type.Type == "client_report" {
		w.Write([]byte("{}"))
		return
	}

	event2 := SentryEvent{}
	err = json.Unmarshal(v[2], &event2)
	if err != nil {
		http.Error(w, "Error parsing event", http.StatusBadRequest)
	}

	if DEBUG {
		fmt.Printf("Received %s event\n", msg_type.Type)
		fmt.Println(string(v[2]))
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
	stacktrace := ""
	if event2.Message != "" {
		message = event2.Message
	} else if event2.LogEntry.Message != "" {
		message = fmt.Sprintf("%s: %s: %s", event2.Logger, event2.LogEntry.Message, event2.LogEntry.Params)
	} else {
		if len(event2.Exception.Values) > 0 {
			// Will it ever be more??
			value := event2.Exception.Values[0]
			message = fmt.Sprintf("%s: %s", value.Type, value.Value)

			journal_metadata["SENTRY_STACKTRACE"], _ = json.Marshal(value.Stacktrace)

			// Prepend the stacktrace
			for _, frame := range value.Stacktrace.Frames {
				stacktrace = fmt.Sprintf("[%s:%d:%d] %s",
					frame.Filename, frame.Lineno, frame.Colno, stacktrace)
			}
		}
	}

	var log_message string
	// What we print to the logs needs to include more useful information
	log_message = fmt.Sprintf("[%s] (proj=%s env=%s) %s %s", msg_type.Type, sentry_key, event2.Environment, stacktrace, message)
	// But what we'll use to de-duplicate, should only include the bare minimum
	journal_metadata["SENTRY_MESSAGE_KEY"] = fmt.Sprintf("(proj=%s) %s", sentry_key, message)
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
