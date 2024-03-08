package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"text/template"

	"github.com/go-chi/chi/v5"
)

func indexPage(w http.ResponseWriter, r *http.Request) {
	list_tpl_text, err := os.ReadFile("list.html")

	entries := GroupMain()

	type Zzz struct {
		Entries []LogEntry
	}

	zzz := Zzz{Entries: entries}

	if err != nil {
		fmt.Println(err)
	}
	list_tpl, err := template.New("list").Parse(string(list_tpl_text))
	if err != nil {
		fmt.Println(err)
	}

	err = list_tpl.Execute(w, zzz)
	if err != nil {
		fmt.Println(err)
	}
}

func issuePage(w http.ResponseWriter, r *http.Request) {
	msg := chi.URLParam(r, "msg")
	// base64 decode
	msg_bytes, err := base64.StdEncoding.DecodeString(msg)

	list_tpl_text, err := os.ReadFile("show.html")
	entries := aggregateIdenticalMessages(string(msg_bytes), 24)

	type Zzy struct {
		Entry   LogEntry
		Message string
	}

	zzz := Zzy{Entry: entries, Message: string(msg_bytes)}

	if err != nil {
		fmt.Println(err)
	}
	list_tpl, err := template.New("show").Parse(string(list_tpl_text))
	if err != nil {
		fmt.Println(err)
	}

	err = list_tpl.Execute(w, zzz)
	if err != nil {
		fmt.Println(err)
	}
}
