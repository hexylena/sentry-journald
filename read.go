package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
	// "github.com/coreos/go-systemd/v22/journal"
	"github.com/coreos/go-systemd/v22/sdjournal"
	"github.com/google/uuid"
	"github.com/mileusna/useragent"
)

type JournalEntry struct {
	E *sdjournal.JournalEntry
}

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

type LogEntry struct {
	LastSeen time.Time
	Entries  []*JournalEntry
	Id       string
}

func (j LogEntry) GetTime() string {
	return j.LastSeen.Format(time.RFC3339)
}

func (e LogEntry) GetMessage() string {
	return e.Entries[0].E.Fields["MESSAGE"]
}

func (e LogEntry) GetId() string {
	m := e.GetMessage()
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

func aggregateIdenticalMessages(msg string, hours int) LogEntry {
	j, err := sdjournal.NewJournal()
	if err != nil {
		panic(err)
	}
	defer j.Close()

	err = j.AddMatch("MESSAGE=" + msg)
	if err != nil {
		panic(err)
	}

	logStart := time.Now().Add(-time.Duration(hours) * time.Hour)
	// dur := time.Duration(hours) * time.Hour
	// timeNow := time.Now().Sub(dur)
	err = j.SeekRealtimeUsec(uint64(logStart.UnixMicro() / 1000))
	if err != nil {
		panic(err)
	}

	entries := LogEntry{}
	for {
		x, err := j.Next()
		if x == 0 {
			return entries
		}
		if err != nil {
			panic(err)
		}
		entry, err := j.GetEntry()
		entries.Entries = append(entries.Entries, &JournalEntry{E: entry})
	}
	return entries
}

func discoverAllMessages(hours int) []*JournalEntry {
	j, err := sdjournal.NewJournal()
	if err != nil {
		panic(err)
	}
	defer j.Close()

	err = j.AddMatch("SYSLOG_IDENTIFIER=sentry")
	if err != nil {
		panic(err)
	}

	logStart := time.Now().Add(-time.Duration(hours) * time.Hour)
	// dur := time.Duration(hours) * time.Hour
	// timeNow := time.Now().Sub(dur)
	err = j.SeekRealtimeUsec(uint64(logStart.UnixMicro() / 1000))
	if err != nil {
		panic(err)
	}

	entries := []*JournalEntry{}
	for {
		x, err := j.Next()
		if x == 0 {
			return entries
		}
		if err != nil {
			panic(err)
		}
		entry, err := j.GetEntry()
		entries = append(entries, &JournalEntry{E: entry})
	}
	return entries
}

func discard() {
	// msg := "[event] (proj=galaxy env=production) [galaxy/webapps/galaxy/controllers/error.py:10:0] [galaxy/web/framework/base.py:262:0] [galaxy/web/framework/base.py:173:0] [/home/user/arbeit/galaxy/galaxy/.venv/lib64/python3.11/site-packages/paste/httpexceptions.py:640:0] [galaxy/web/framework/middleware/remoteuser.py:201:0] [galaxy/web/framework/middleware/error.py:165:0] [a2wsgi/wsgi.py:198:0] [a2wsgi/wsgi.py:157:0] [starlette/routing.py:443:0] [starlette/routing.py:718:0] [fastapi/middleware/asyncexitstack.py:17:0] [fastapi/middleware/asyncexitstack.py:20:0] [starlette/middleware/exceptions.py:68:0] [starlette/middleware/exceptions.py:79:0] [starlette/middleware/base.py:70:0] [starlette/middleware/base.py:98:0] [starlette/responses.py:262:0] [starlette/middleware/base.py:134:0] [starlette/responses.py:273:0] Exception: Fake error"
	msg := "[event] (proj=webdemo env=production) My Message"

	entries := aggregateIdenticalMessages(msg, 9)
	for _, entry := range entries.Entries {
		header_map := make(map[string]string)
		json.Unmarshal([]byte(entry.E.Fields["REQUEST_HEADERS"]), &header_map)
		ua := useragent.Parse(header_map["User-Agent"])
		fmt.Println(ua.OS, ua.OSVersion, ua.Name, ua.Version, ua.Desktop, ua.Mobile, ua.Bot, ua.Device)
		fmt.Println(entry.E.Fields)
	}
}

func GroupMain() []LogEntry {

	// func main(){
	//
	entries := discoverAllMessages(1)
	grouped := make(map[string]LogEntry)

	for _, entry := range entries {
		k := entry.E.Fields["MESSAGE"]

		ref, found := grouped[k]
		if !found {
			grouped[k] = LogEntry{
				LastSeen: time.UnixMicro(int64(entry.E.RealtimeTimestamp)),
				Entries:  []*JournalEntry{},
				Id:       uuid.New().String(),
			}
		}

		ref = grouped[k]
		ref.Entries = append(ref.Entries, entry)
		ut := time.UnixMicro(int64(entry.E.RealtimeTimestamp))
		if ut.After(ref.LastSeen) {
			ref.LastSeen = ut
		}
		grouped[k] = ref
	}

	// get the keys, sorted
	keys := make([]LogEntry, 0, len(grouped))
	for _, v := range grouped {
		keys = append(keys, v)
	}
	// Sort
	sort.Sort(ByAge(keys))

	return keys
}
