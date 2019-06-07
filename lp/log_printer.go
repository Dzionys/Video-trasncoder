package lp

import (
	"log"
	"os"
	"strings"
)

var LogFile *os.File

func WLog(msg string, clid string) {
	if strings.HasPrefix(msg, "Error:") {
		log.Println(msg)
	}
	UpdateLogMessage(msg, clid)
}

func OpenLogFile(filepath string) error {
	var err error

	// Create log file if not exist
	if _, err = os.Stat(filepath); os.IsNotExist(err) {
		_, err = os.Create(filepath)
		if err != nil {
			log.Fatalf("error creating file: %v", err)
			return err
		}
	}

	// Open log file
	LogFile, err = os.OpenFile("info.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
		return err
	}

	// Set log writer to log file insted of std
	log.SetOutput(LogFile)

	return nil
}
