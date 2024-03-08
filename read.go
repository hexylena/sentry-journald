package main

import (
	"sort"
	"time"
	// "github.com/coreos/go-systemd/v22/journal"
	"github.com/coreos/go-systemd/v22/sdjournal"
	"github.com/google/uuid"
)

// A small wrapper around sdjournal so we can define our own functions on it.
type JournalEntry struct {
	E *sdjournal.JournalEntry
}

func aggregateIdenticalMessages(msg string, hours int) LogEntry {
	j, err := sdjournal.NewJournal()
	if err != nil {
		panic(err)
	}
	defer j.Close()

	err = j.AddMatch("SENTRY_MESSAGE_KEY=" + msg)
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

func GroupMain() []LogEntry {

	// func main(){
	//
	entries := discoverAllMessages(1)
	grouped := make(map[string]LogEntry)

	for _, entry := range entries {
		k := entry.E.Fields["SENTRY_MESSAGE_KEY"]

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
