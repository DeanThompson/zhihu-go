package zhihu

import (
	"testing"
)

func Test_validQuestionURL(t *testing.T) {
	io_map := map[string]bool{
		"https://www.zhihu.com/question/37284137":  true,
		"http://www.zhihu.com/question/41114729":   true,
		"https://www.zhihu.com/question/41114729x": false,
		"https://www.zhihu.com/question/4111472":   false,
		"https://www.zhihu.com/":                   false,
	}

	for value, expectedResult := range io_map {
		if validQuestionURL(value) != expectedResult {
			t.Error("validQuestionURL returns error result")
		}
	}
}

func Test_openCaptchaFile(t *testing.T) {
	err := openCaptchaFile("./examples/verify.gif")
	if err != nil {
		t.Errorf("openCaptchaFile fails: %s", err.Error())
	}
}
