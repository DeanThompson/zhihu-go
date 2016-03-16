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
	questionURLPattern   = regexp.MustCompile("^(http|https)://www.zhihu.com/question/[0-9]{8}$")
	collectionURLPattern = regexp.MustCompile("^(http|https)://www.zhihu.com/collection/[0-9]{8}$")
	logger               = Logger{Enabled: true}
)

func validQuestionURL(value string) bool {
	return questionURLPattern.MatchString(value)
}

func validCollectionURL(value string) bool {
	return collectionURLPattern.MatchString(value)
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

func minInt(a, b int) int {
	if a > b {
		return b
	}
	return a
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

// newDocumentFromUrl 会请求给定的 url，并返回一个 goquery.Document 对象用于解析
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

// ZhihuPage 是一个知乎页面，User, Question, Answer, Collection 的公共部分
type ZhihuPage struct {
	// Link 是该页面的链接
	Link string

	// doc 是 HTML document
	doc *goquery.Document

	// fields 是字段缓存，避免重复解析页面
	fields map[string]interface{}
}

// newZhihuPage 是 private 的构造器
func newZhihuPage(link string) *ZhihuPage {
	return &ZhihuPage{
		Link:   link,
		fields: make(map[string]interface{}),
	}
}

// Doc 用于获取当前问题页面的 HTML document，惰性求值
func (page *ZhihuPage) Doc() *goquery.Document {
	if page.doc != nil {
		return page.doc
	}

	err := page.Refresh()
	if err != nil {
		return nil
	}

	return page.doc
}

// Refresh 会重新载入当前页面，获取最新的数据
func (page *ZhihuPage) Refresh() (err error) {
	page.fields = make(map[string]interface{})    // 清空缓存
	page.doc, err = newDocumentFromUrl(page.Link) // 重载页面
	return err
}

// GetXsrf 从当前页面内容抓取 xsrf 的值
func (page *ZhihuPage) GetXsrf() string {
	doc := page.Doc()
	value, _ := doc.Find(`input[name="_xsrf"]`).Attr("value")
	return value
}
