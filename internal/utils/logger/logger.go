package logger

import (
	"encoding/json"
	"log"
	"time"
)

type Entry map[string]interface{}

func Info(msg string, fields Entry) {
	emit("info", msg, fields)
}

func Warn(msg string, fields Entry) {
	emit("warn", msg, fields)
}

func Error(msg string, fields Entry) {
	emit("error", msg, fields)
}

func emit(level, msg string, fields Entry) {
	if fields == nil {
		fields = Entry{}
	}
	fields["level"] = level
	fields["msg"] = msg
	fields["ts"] = time.Now().UTC().Format(time.RFC3339Nano)
	b, err := json.Marshal(fields)
	if err != nil {
		log.Printf("[%s] %s %+v", level, msg, fields)
		return
	}
	log.Print(string(b))
}
