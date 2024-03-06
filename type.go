package main

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

type SentryLogEntry struct {
	Message string   `json:"message"`
	Params  []string `json:"params"`
}

type SentryEvent struct {
	Message    string                 `json:"message"`
	Level      string                 `json:"level"`
	Logger     string                 `json:"logger"`
	LogEntry   SentryLogEntry         `json:"logentry"`
	Event_id   string                 `json:"event_id"`
	Timestamp  interface{}            `json:"timestamp"`
	Contexts   map[string]interface{} `json:"contexts"`
	Extra      map[string]interface{} `json:"extra"`
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
