package zhihu

import (
	"testing"
)

func Test_Error(t *testing.T) {
	var logger = Logger{Enabled: true}
	logger.Error("测试：输出一条 ERROR 的信息")
	logger.Error("测试：从 1 到 5 分别是：%d, %d, %d, %d, %d", 1, 2, 3, 4, 5)
}

func Test_Info(t *testing.T) {
	var logger = Logger{Enabled: true}
	logger.Info("测试：输出一条 INFO 的信息")
	logger.Info("测试：从 1 到 5 分别是：%d, %d, %d, %d, %d", 1, 2, 3, 4, 5)
}
