package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
)

type logLevelInfo struct {
	Color  *color.Color
	Prefix string
}

type logLevel int

const (
	ll_DEBUG logLevel = iota
	ll_SUCCESS
	ll_WARN
	ll_ERROR
	ll_FATAL
)

var (
	debugMention = ""
	debugWebhook = ""
)

var file *os.File
var fileLock = &sync.Mutex{}

var levelInfo = map[logLevel]logLevelInfo{
	ll_DEBUG: {
		Prefix: "DEBUG",
		Color:  color.New(color.FgCyan),
	},
	ll_SUCCESS: {
		Prefix: "SUCCESS",
		Color:  color.New(color.FgGreen),
	},
	ll_WARN: {
		Prefix: "WARNING",
		Color:  color.New(color.FgYellow),
	},
	ll_ERROR: {
		Prefix: "ERROR",
		Color:  color.New(color.FgRed),
	},
	ll_FATAL: {
		Prefix: "FATAL ERROR",
		Color:  color.New(color.FgWhite, color.BgRed),
	},
}

func Init(DEBUG_MENTION, DEBUG_WEBHOOK string) {
	f, err := os.Create("log.txt")
	if err != nil {
		panic(err)
	}
	file = f
	debugMention = DEBUG_MENTION
	debugWebhook = DEBUG_WEBHOOK

	Success("Logger initiated!")
}

type discordWebhook struct {
	Content string `json:"content"`
}

func log(level logLevel, msg string, args ...any) {
	levelInfo := levelInfo[level]

	msg = fmt.Sprintf(msg, args...)

	date := time.Now().Format(`02 Jan 2006 15:04:05`)
	prefix := fmt.Sprintf("[%v] [%v] ", levelInfo.Prefix, date)

	levelInfo.Color.Println(prefix + msg)

	fileLock.Lock()
	file.WriteString(prefix + msg + "\n")
	fileLock.Unlock()

	if level > ll_SUCCESS && debugWebhook != "" {
		content := prefix + "\n```\n" + msg + "```"
		if debugMention != "" {
			content = debugMention + ", " + content
		}

		buf, _ := json.Marshal(discordWebhook{
			Content: content,
		})

		http.Post(debugWebhook, "application/json", bytes.NewReader(buf))
	}
}

func Debug(msg string, args ...any) {
	log(ll_DEBUG, msg, args...)
}

func Success(msg string, args ...any) {
	log(ll_SUCCESS, msg, args...)
}

func Warn(msg string, args ...any) {
	log(ll_WARN, msg, args...)
}

func Error(msg string, args ...any) {
	log(ll_ERROR, msg, args...)
}

func Fatal(msg string, args ...any) {
	log(ll_FATAL, msg, args...)
	os.Exit(1)
}

// "While {CONTEXT}: {ERROR}"
func FatalIfErr(err error, context string, args ...any) {
	if err != nil {
		context = fmt.Sprintf(context, args...)
		Fatal("Error while %s: %s", context, err.Error())
	}
}

// "While {CONTEXT}: {ERROR}"
// Returns true if err != nil, intended for inline usage:
// if ErrorIfErr(err, "context here (%s)", "stuff") { return }
func ErrorIfErr(err error, context string, args ...any) bool {
	if err != nil {
		context = fmt.Sprintf(context, args...)

		Error("Error while %s: %s", context, err.Error())
		return true
	}
	return false
}

// Doesn't log anything, but prints in the pretty debug color
func DebugPretty(msg string, args ...any) {
	levelInfo[ll_DEBUG].Color.Printf(msg+"\n", args...)
}

// Doesn't log anything, but prints in the pretty success color
func SuccessPretty(msg string, args ...any) {
	levelInfo[ll_SUCCESS].Color.Printf(msg+"\n", args...)
}

func Close() {
	file.Close()
}
