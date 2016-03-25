package zhihu

import (
	"fmt"

	"github.com/fatih/color"
)

// Logger 是一个简单的输出工具，可以输出不同颜色的信息
// TODO simple level
type Logger struct {
	Enabled bool
}

func (logger *Logger) log(a ...interface{}) {
	if logger.Enabled {
		fmt.Println(a...)
	}
}

// Error 输出 error 级别的日志
func (logger *Logger) Error(msg string, a ...interface{}) {
	logger.log(color.RedString("ERROR: "+msg, a...))
}

// Warn 输出 warning 级别的日志
func (logger *Logger) Warn(msg string, a ...interface{}) {
	logger.log(color.YellowString("WARN: "+msg, a...))
}

// Warning 是 Warn 的别名
func (logger *Logger) Warning(msg string, a ...interface{}) {
	logger.Warn(msg, a...)
}

// Info 输出 info 级别的日志
func (logger *Logger) Info(msg string, a ...interface{}) {
	logger.log(color.BlueString("INFO: "+msg, a...))
}

// Debug 输出 debug 级别的日志
func (logger *Logger) Debug(msg string, a ...interface{}) {
	logger.log(color.WhiteString("DEBUG: "+msg, a...))
}

// Success 输出 success 的日志，基本上与 info 一样，除了使用了绿色
func (logger *Logger) Success(msg string, a ...interface{}) {
	logger.log(color.GreenString("SUCCESS: "+msg, a...))
}
