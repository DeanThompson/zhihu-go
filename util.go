package zhihu

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
)

const (
	userAgent    = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.116 Safari/537.36"
	baseZhihuURL = "https://www.zhihu.com"
	pageSize     = 20
)

var (
	reQuestionURL   = regexp.MustCompile("^(http|https)://www.zhihu.com/question/[0-9]{8}$")
	reCollectionURL = regexp.MustCompile("^(http|https)://www.zhihu.com/collection/[0-9]{8}$")
	reTopicURL      = regexp.MustCompile("^(http|https)://www.zhihu.com/topic/[0-9]{8}$")
	reGetNumber     = regexp.MustCompile(`([0-9])+`)
	logger          = Logger{Enabled: true}
)

func validQuestionURL(value string) bool {
	return reQuestionURL.MatchString(value)
}

func validCollectionURL(value string) bool {
	return reCollectionURL.MatchString(value)
}

func validTopicURL(value string) bool {
	return reTopicURL.MatchString(value)
}

func reMatchInt(raw string) int {
	matched := reGetNumber.FindStringSubmatch(raw)
	if len(matched) == 0 {
		return 0
	}
	rv, _ := strconv.Atoi(matched[0])
	return rv
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

func save(filename string, content []byte) error {
	return ioutil.WriteFile(filename, content, 0666)
}

func saveString(filename string, content string) error {
	return ioutil.WriteFile(filename, []byte(content), 0666)
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
	return urlJoin(baseZhihuURL, path)
}

func urlJoin(base, path string) string {
	if strings.HasSuffix(base, "/") {
		base = strings.TrimRight(base, "/")
	}
	if strings.HasPrefix(path, "/") {
		path = strings.TrimLeft(path, "/")
	}
	return base + "/" + path
}

// newDocumentFromUrl 会请求给定的 url，并返回一个 goquery.Document 对象用于解析
func newDocumentFromURL(url string) (*goquery.Document, error) {
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
type Page struct {
	// Link 是该页面的链接
	Link string

	// doc 是 HTML document
	doc *goquery.Document

	// fields 是字段缓存，避免重复解析页面
	fields map[string]interface{}
}

// newZhihuPage 是 private 的构造器
func newZhihuPage(link string) *Page {
	return &Page{
		Link:   link,
		fields: make(map[string]interface{}),
	}
}

// Doc 用于获取当前问题页面的 HTML document，惰性求值
func (page *Page) Doc() *goquery.Document {
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
func (page *Page) Refresh() (err error) {
	page.fields = make(map[string]interface{})    // 清空缓存
	page.doc, err = newDocumentFromURL(page.Link) // 重载页面
	return err
}

// GetXsrf 从当前页面内容抓取 xsrf 的值
func (page *Page) GetXSRF() string {
	doc := page.Doc()
	value, _ := doc.Find(`input[name="_xsrf"]`).Attr("value")
	return value
}

func (page *Page) setField(field string, value interface{}) {
	page.fields[field] = value
}

func (page *Page) getIntField(field string) (value int, exists bool) {
	if got, ok := page.fields[field]; ok {
		return got.(int), true
	}
	return 0, false
}

func (page *Page) getStringField(field string) (value string, exists bool) {
	if got, ok := page.fields[field]; ok {
		return got.(string), true
	}
	return "", false
}

// nodeListResult 是形如 /node/XXListV2 这样的 Ajax 请求的 JSON 返回值
type nodeListResult struct {
	R   int      `json:"r"`   // 状态码，正确的情况为 0
	Msg []string `json:"msg"` // 回答内容，每个元素都是一段 HTML 片段
}

// normalAjaxResult 是页面内，目标 URL 和当前页面 URL 相同的 Ajax 请求返回的 JSON 数据
type normalAjaxResult struct {
	R   int           `json:"r"`
	Msg []interface{} `json:"msg"` // 两个元素，第一个为话题数量，第二个是 HTML 片段
}
