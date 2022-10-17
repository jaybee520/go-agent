package util

import (
	"fmt"
)
import "time"

//import "david/util"

var StdLogger = &Logger{
	level: -1,
}

type Logger struct {
	level int //1个level对应4个空格
}

//LevelUp
func (logger *Logger) LevelUp() {
	logger.level += 1
}

//LevelDown
func (logger *Logger) LevelDown() {
	logger.level -= 1
	if logger.level < 0 {
		fmt.Errorf("logger.level should not less than 0 %d\n", logger.level)
	}
}

//Debug print debug level
func (logger *Logger) Debug(format string, a ...interface{}) {
	//logger.Debug(format, a)

}

//Info print debug level
func (logger *Logger) Info(format string, a ...interface{}) {
	//logger.Info(format, a)
}

//Warn print debug level
func (logger *Logger) Warn(format string, a ...interface{}) {
	//logger.Warn(format, a)
}

//Error print debug level
func (logger *Logger) Error(format string, a ...interface{}) {
	//logger.Error(format, a)
}

//nowTime hh:mm:ss
func NowTime() string {
	now := time.Now()
	return fmt.Sprintf("%02d:%02d:%02d", now.Hour(), now.Minute(), now.Second())
}
