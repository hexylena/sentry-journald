package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	// "github.com/coreos/go-systemd/v22/journal"
	"github.com/mileusna/useragent"
)

func (j JournalEntry) GetField(key string) string {
	return j.E.Fields[key]
}

func (j JournalEntry) GetTime() string {
	return time.UnixMicro(int64(j.E.RealtimeTimestamp)).Format(time.RFC3339)
}

func (j JournalEntry) GetTimeTime() time.Time {
	return time.UnixMicro(int64(j.E.RealtimeTimestamp))
}

func (j JournalEntry) GetLogLevel() string {
	switch j.GetField("PRIORITY") {
	case "0":
		return "Emergency"
	case "1":
		return "Alert"
	case "2":
		return "Critical"
	case "3":
		return "Error"
	case "4":
		return "Warning"
	case "5":
		return "Notice"
	case "6":
		return "Informational"
	case "7":
		return "Debug"
	}
	return "UNKNOWN"
}

// These are a collection of journal entries.
type LogEntry struct {
	LastSeen time.Time
	Entries  []*JournalEntry
	Id       string
}

func (j LogEntry) EntriesReverse() []*JournalEntry {
	var entries []*JournalEntry
	for i := len(j.Entries) - 1; i >= 0; i-- {
		entries = append(entries, j.Entries[i])
	}
	return entries
}

func (j LogEntry) GetTime() string {
	return j.LastSeen.Format(time.RFC3339)
}

func (e LogEntry) GetProject() string {
	return e.Entries[0].E.Fields["SENTRY_KEY"]
}

func (e LogEntry) GetMessage() string {
	return e.Entries[0].E.Fields["MESSAGE"]
}

func (e LogEntry) GetMessageKey() string {
	return e.Entries[0].E.Fields["SENTRY_MESSAGE_KEY"]
}

func (e LogEntry) HasStacktrace() bool {
	_, ok := e.Entries[0].E.Fields["SENTRY_STACKTRACE"]
	return ok
}

func (e LogEntry) GetStacktrace() SentryStackTrace {
	var s SentryStackTrace
	json.Unmarshal([]byte(e.Entries[0].E.Fields["SENTRY_STACKTRACE"]), &s)
	return s
}

func (e LogEntry) GetId() string {
	m := e.GetMessageKey()
	// base64 encode this:
	return base64.StdEncoding.EncodeToString([]byte(m))
}

func (e LogEntry) GetBrowserMeta(attr string) map[string]float64 {
	browsers := make(map[string]int)
	for _, entry := range e.Entries {
		header_map := make(map[string]string)
		json.Unmarshal([]byte(entry.E.Fields["REQUEST_HEADERS"]), &header_map)
		// lowercase every key in header_map
		for k, v := range header_map {
			if k != strings.ToLower(k) {
				header_map[strings.ToLower(k)] = v
				delete(header_map, k)
			}
		}
		ua := useragent.Parse(header_map["user-agent"])

		var k string
		switch attr {
		case "os":
			k = fmt.Sprintf("%s %s", ua.OS, ua.OSVersion)
		case "version":
			k = fmt.Sprintf("%s %s", ua.Name, ua.Version)
		case "browser":
			k = ua.Name
		case "device":
			k = ua.Device
		}

		if _, ok := browsers[k]; ok {
			browsers[k] = browsers[k] + 1
		} else {
			browsers[k] = 1
		}
	}

	// convert to percentages
	total := len(e.Entries)
	results := make(map[string]float64)
	for k, v := range browsers {
		if k == "" || k == " " {
			continue
		}
		results[k] = float64(v) / float64(total)
	}

	fmt.Println(results)
	return results
}

func (e LogEntry) GetMeta(attr string) map[string]float64 {
	browsers := make(map[string]int)
	for _, entry := range e.Entries {
		var k = entry.E.Fields[attr]

		if _, ok := browsers[k]; ok {
			browsers[k] = browsers[k] + 1
		} else {
			browsers[k] = 1
		}
	}

	// convert to percentages
	total := len(e.Entries)
	results := make(map[string]float64)
	for k, v := range browsers {
		if k == "" {
			continue
		}
		fmt.Printf("k=%s v=%s.\n", k, v)
		results[k] = float64(v) / float64(total)
	}

	return results
}

func (e LogEntry) GetCount() int {
	return len(e.Entries)
}

func (e LogEntry) GetLogLevel() string {
	return e.Entries[0].GetLogLevel()
}

func (e LogEntry) GetHistogram12h() []float64 {
	// 12 hours
	hist := make([]int, 12)
	m := 0
	for _, entry := range e.Entries {
		// hours ago
		hour := int(time.Now().Sub(entry.GetTimeTime()).Hours())
		if hour > 11 {
			continue
		}

		hist[11-hour] = hist[11-hour] + 1
		if hist[11-hour] > m {
			m = hist[11-hour]
		}
	}

	results := make([]float64, 12)
	for i, v := range hist {
		results[i] = float64(v) / float64(m)
	}

	return results
}

type ByAge []LogEntry

func (a ByAge) Len() int           { return len(a) }
func (a ByAge) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByAge) Less(i, j int) bool { return a[i].LastSeen.After(a[j].LastSeen) }
