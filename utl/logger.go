package utl

import (
	echoLog "github.com/labstack/gommon/log"
	"os"
)

func GetLogger(prefix, path string, level echoLog.Lvl) *echoLog.Logger {
	logger := echoLog.New(prefix)
	logger.SetLevel(level)
	file, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, os.ModePerm)
	if err != nil {
		logger.Warnf("%v, set logger output to file fail, err:%v", "", err)
		logger.SetOutput(os.Stdout)
	} else {
		logger.SetOutput(file)
	}
	// set header
	header := `${time_rfc3339}  ${level}  file:${short_file}:${line}  prefix:${prefix}  msg:`
	logger.SetHeader(header)
	return logger
}

type Logger interface {
}
