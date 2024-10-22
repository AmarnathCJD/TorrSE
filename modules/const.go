package modules

import (
	"log"
)

func getUA() string {
	return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36"
}

var engineLogger = log.New(log.Writer(), "engine: ", log.LstdFlags)

func logInfo(msg string) {
	engineLogger.Println(msg)
}

func logError(err error) {
	engineLogger.Println(err)
}

func LogGlobalInfo(msg string) {
	logInfo(msg)
}

func LogGlobalError(err error) {
	logError(err)
}
