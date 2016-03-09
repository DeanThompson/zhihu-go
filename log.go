package zhihu

import (
	"fmt"

	"github.com/fatih/color"
)

// Logger 是一个简单的输出工具，可以输出不同颜色的信息
type Logger struct {
	Enabled bool
}

func (logger *Logger) log(a ...interface{}) {
	if logger.Enabled {
		fmt.Println(a...)
	}
}

func (logger *Logger) Error(msg string, a ...interface{}) {
	logger.log(color.RedString("ERROR:"+msg, a...))
}

func (logger *Logger) Warn(msg string, a ...interface{}) {
	logger.log(color.YellowString("WARN: "+msg, a...))
}

func (logger *Logger) Warning(msg string, a ...interface{}) {
	logger.Warn(msg, a...)
}

func (logger *Logger) Info(msg string, a ...interface{}) {
	logger.log(color.BlueString("INFO: "+msg, a...))
}

func (logger *Logger) Debug(msg string, a ...interface{}) {
	logger.log(color.WhiteString("DEBUG: "+msg, a...))
}

func (logger *Logger) Success(msg string, a ...interface{}) {
	logger.log(color.GreenString("SUCCESS: "+msg, a...))
}
