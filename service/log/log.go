package log

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

var LogLevel int = 1
var wg sync.WaitGroup

const (
	reset  = "\033[0m"
	purple = "\033[35m"
	gray   = "\033[90m"
	blue   = "\033[34m"
	yellow = "\033[33m"
	red    = "\033[31m"
	green  = "\033[32m"
)

type logMsg struct {
	level int
	color string
	tag   string
	msg   string
}

var logChan = make(chan logMsg, 125)

func enqueue(level int, color, tag string, format any, a ...any) {
	var message string

	switch v := format.(type) {
	case string:
		if len(a) > 0 {
			message = fmt.Sprintf(v, a...)
		} else {
			message = v
		}

	default:
		message = fmt.Sprint(format)
	}

	logChan <- logMsg{level: level, color: color, tag: tag, msg: message}
}

func Trace(format any, a ...any) {
	enqueue(0, purple, "TRACE", format, a...)
}

func Debug(format any, a ...any) {
	enqueue(1, gray, "DEBUG", format, a...)
}

func Info(format any, a ...any) {
	enqueue(2, blue, "INFO", format, a...)
}

func Warn(format any, a ...any) {
	enqueue(3, yellow, "WARN", format, a...)
}

func Error(format any, a ...any) {
	enqueue(4, red, "ERROR", format, a...)
}

func Done(format any, a ...any) {
	enqueue(5, green, "DONE", format, a...)
}

func Print(format any, a ...any) {
	enqueue(6, reset, " LOG ", format, a...)
}

func Shutdown() {
	close(logChan)
	wg.Wait()
}

func init() {
	wg.Go(func() {
		for msg := range logChan {
			if LogLevel <= msg.level {
				utc := time.Now().UTC()
				timeStamp := fmt.Sprintf("%s UTC", utc.Format(time.RFC3339))
				level := fmt.Sprintf("%s| %s |", msg.color, msg.tag)

				fmt.Println(timeStamp, level, msg.msg, reset)
			}
		}
	})

	if level, done := os.LookupEnv("LOG_LEVEL"); done {
		val, err := strconv.Atoi(level)
		if err == nil {
			LogLevel = val
		} else {
			LogLevel = 2
		}
	} else {
		LogLevel = 1
	}
}
