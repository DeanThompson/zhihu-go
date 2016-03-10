package zhihu

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
)

const (
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.116 Safari/537.36"
)

var (
	questionURLPattern = regexp.MustCompile("^(http|https)://www.zhihu.com/question/[0-9]{8}$")
	logger             = Logger{Enabled: true}
)

func validQuestionURL(value string) bool {
	return questionURLPattern.Match([]byte(value))
}

func newHTTPHeaders(isXhr bool) http.Header {
	headers := make(http.Header)
	headers.Set("Accept", "*/*")
	headers.Set("Connection", "keep-alive")
	headers.Set("Host", "www.zhihu.com")
	headers.Set("Origin", "http://www.zhihu.com")
	headers.Set("Pragma", "no-cache")
	headers.Set("User-Agent", userAgent)
	if isXhr {
		headers.Set("X-Requested-With", "XMLHttpRequest")
	}
	return headers
}

func strip(s string) string {
	return strings.Trim(s, "\n")
}

func getCwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		panic("获取 CWD 失败：" + err.Error())
	}
	return cwd
}

func openCaptchaFile(filename string) error {
	logger.Info("调用外部程序渲染验证码……")
	var args []string
	switch runtime.GOOS {
	case "linux":
		args = []string{"xdg-open", filename}
	case "darwin":
		args = []string{"open", filename}
	case "freebsd":
		args = []string{"open", filename}
	case "netbsd":
		args = []string{"open", filename}
	case "windows":
		var (
			cmd      = "url.dll,FileProtocolHandler"
			runDll32 = filepath.Join(os.Getenv("SYSTEMROOT"), "System32", "rundll32.exe")
		)
		args = []string{runDll32, cmd, filename}
	default:
		fmt.Printf("无法确定操作系统，请自行打开验证码 %s 文件，并输入验证码。", filename)
	}

	logger.Info("Command: %s", strings.Join(args, " "))

	err := exec.Command(args[0], args[1:]...).Run()
	if err != nil {
		return err
	}

	return nil
}

func readCaptchaInput() string {
	var captcha string
	fmt.Print(color.CyanString("请输入验证码："))
	fmt.Scanf("%s", &captcha)
	return captcha
}

func makeZhihuLink(path string) string {
	path = strings.TrimLeft(path, "/")
	return "http://www.zhihu.com/" + path
}

func newDocumentFromUrl(url string) (*goquery.Document, error) {
	resp, err := gSession.Get(url)
	if err != nil {
		logger.Error("请求 %s 失败：%s", url, err.Error())
		return nil, err
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		logger.Error("解析页面失败：%s", err.Error())
	}

	return doc, err
}
