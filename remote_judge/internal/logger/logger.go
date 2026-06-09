package logger

import (
	"encoding/json"
	"log"
	"time"
)

// Event 是一条结构化日志条目。
type Event struct {
	Time    string         `json:"time"`
	Level   string         `json:"level"`
	Name    string         `json:"name"`
	TraceID string         `json:"traceId,omitempty"`
	Message string         `json:"message,omitempty"`
	Fields  map[string]any `json:"fields,omitempty"`
}

// Info 写入一条结构化 info 级别日志。
func Info(name, traceID, message string, fields map[string]any) {
	write("INFO", name, traceID, message, fields)
}

// Error 写入一条结构化 error 级别日志。
func Error(name, traceID, message string, fields map[string]any) {
	write("ERROR", name, traceID, message, fields)
}

// write serializes a structured log entry as JSON.
func write(level, name, traceID, message string, fields map[string]any) {
	entry := Event{
		Time:    time.Now().Format(time.RFC3339Nano),
		Level:   level,
		Name:    name,
		TraceID: traceID,
		Message: message,
		Fields:  fields,
	}
	if payload, err := json.Marshal(entry); err == nil {
		log.Print(string(payload))
		return
	}
	log.Printf(`{"time":"%s","level":"%s","name":"%s","traceId":"%s","message":"%s"}`, entry.Time, level, name, traceID, message)
}
