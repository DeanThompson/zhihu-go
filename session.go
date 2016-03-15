package zhihu

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/juju/persistent-cookiejar"
)

type Auth struct {
	Account  string `json:"account"`
	Password string `json:"password"`

	loginType string // phone_num 或 email
	loginUrl  string // 通过 Account 判断
}

// isEmail 判断是否通过邮箱登录
func (auth *Auth) isEmail() bool {
	return regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`).MatchString(auth.Account)
}

// isPhone 判断是否通过手机号登录
func (auth *Auth) isPhone() bool {
	return regexp.MustCompile(`^1[0-9]{10}$`).MatchString(auth.Account)
}

func (auth *Auth) toForm() url.Values {
	if auth.isEmail() {
		auth.loginType = "email"
		auth.loginUrl = "http://www.zhihu.com/login/email"
	} else if auth.isPhone() {
		auth.loginType = "phone_num"
		auth.loginUrl = "http://www.zhihu.com/login/phone_num"
	} else {
		panic("无法判断登录类型: " + auth.Account)
	}
	values := url.Values{}
	logger.Info("登录类型：%s, 登录地址：%s", auth.loginType, auth.loginUrl)
	values.Set(auth.loginType, auth.Account)
	values.Set("password", auth.Password)
	values.Set("remember_me", "true") // import!
	return values
}

type Session struct {
	auth   *Auth
	client *http.Client
}

type loginResult struct {
	R         int         `json:"r"`
	Msg       string      `json:"msg"`
	ErrorCode int         `json:"errcode"`
	Data      interface{} `json:"data"`
}

// NewSession 创建并返回一个 *Session 对象，
// 这里没有初始化登录账号信息，账号信息用 `LoadConfig` 通过配置文件进行设置
func NewSession() *Session {
	s := new(Session)
	cookieJar, _ := cookiejar.New(nil)
	s.client = &http.Client{
		Jar: cookieJar,
	}
	return s
}

// LoadConfig 从配置文件中读取账号信息
// 配置文件 是 JSON 格式：
// {
//   "account": "xyz@example.com",
//   "password": "p@ssw0rd"
// }
func (s *Session) LoadConfig(cfg string) {
	fd, err := os.Open(cfg)
	if err != nil {
		panic("无法打开配置文件 config.json: " + err.Error())
	}
	defer fd.Close()

	auth := new(Auth)
	err = json.NewDecoder(fd).Decode(&auth)
	if err != nil {
		panic("解析配置文件出错: " + err.Error())
	}

	s.auth = auth
	// TODO 如果设置了与上一次不一样的账号，最好把 cookies 重置
}

// Login 登录并保存 cookies
func (s *Session) Login() error {
	if s.authenticated() {
		logger.Success("已经是登录状态，不需要重复登录")
		return nil
	}

	form := s.buildLoginForm().Encode()
	body := strings.NewReader(form)
	req, err := http.NewRequest("POST", s.auth.loginUrl, body)
	if err != nil {
		logger.Error("构造登录请求失败：%s", err.Error())
		return err
	}

	headers := newHTTPHeaders(true)
	headers.Set("Content-Length", strconv.Itoa(len(form)))
	headers.Set("Content-Type", "application/x-www-form-urlencoded")
	headers.Set("Referer", "http://www.zhihu.com/")
	req.Header = headers

	logger.Info("登录中，用户名：%s", s.auth.Account)

	resp, err := s.client.Do(req)
	if err != nil {
		logger.Error("登录失败：%s", err.Error())
		return err
	}

	if strings.ToLower(resp.Header.Get("Content-Type")) != "application/json" {
		logger.Error("服务器没有返回 json 数据")
		return fmt.Errorf("未知的 Content-Type: %s", resp.Header.Get("Content-Type"))
	}

	defer resp.Body.Close()
	result := loginResult{}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error("读取响应内容失败：%s", err.Error())
	}

	logger.Info("登录响应内容：%s", strings.Replace(string(content), "\n", "", -1))

	err = json.Unmarshal(content, &result)
	if err != nil {
		logger.Error("JSON 解析失败：%s", err.Error())
		return err
	}

	if result.R == 0 {
		logger.Success("登录成功！")
		s.client.Jar.(*cookiejar.Jar).Save()
		return nil
	}
	if result.R == 1 {
		logger.Warn("登录失败！原因：%s", result.Msg)
		return fmt.Errorf("登录失败！原因：%s", result.Msg)
	}

	logger.Error("登录出现未知错误：%s", string(content))
	return fmt.Errorf("登录失败，未知错误：%s", string(content))
}

// Get 发起一个 GET 请求，自动处理 cookies
func (s *Session) Get(url string) (*http.Response, error) {
	logger.Info("GET %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Error("NewRequest failed with URL: %s", url)
		return nil, err
	}

	req.Header = newHTTPHeaders(false)
	return s.client.Do(req)
}

// Post 发起一个 POST 请求，自动处理 cookies
func (s *Session) Post(url string, bodyType string, body io.Reader) (*http.Response, error) {
	logger.Info("POST %s, %s", url, bodyType)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	headers := newHTTPHeaders(false)
	headers.Set("Content-Type", bodyType)
	req.Header = headers
	return s.client.Do(req)
}

// authenticated 检查是否已经登录（cookies 没有失效）
func (s *Session) authenticated() bool {
	originUrl := "https://www.zhihu.com/settings/profile"
	resp, err := s.Get(originUrl)
	if err != nil {
		logger.Error("访问 profile 页面出错: %s", err.Error())
		return false
	}

	// 如果没有登录，会跳转到 http://www.zhihu.com/?next=%2Fsettings%2Fprofile
	lastUrl := resp.Request.URL.String()
	logger.Info("获取 profile 的请求，跳转到了：%s", lastUrl)
	return lastUrl == originUrl
}

func (s *Session) buildLoginForm() url.Values {
	values := s.auth.toForm()
	values.Set("_xsrf", s.searchXsrf())
	values.Set("captcha", s.downloadCaptcha())
	return values
}

// 从 cookies 获取 _xsrf 用于 POST 请求
func (s *Session) searchXsrf() string {
	resp, err := s.Get("http://www.zhihu.com/")
	if err != nil {
		panic("获取 _xsrf 失败：" + err.Error())
	}

	// retrieve from cookies
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "_xsrf" {
			return cookie.Value
		}
	}

	return ""
}

// downloadCaptcha 获取验证码，用于登录
func (s *Session) downloadCaptcha() string {
	url := fmt.Sprintf("http://www.zhihu.com/captcha.gif?r=%d", time.Now().Unix())
	logger.Info("获取验证码：%s", url)
	resp, err := s.Get(url)
	if err != nil {
		panic("获取验证码失败：" + err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("获取验证码失败，StatusCode = %d", resp.StatusCode))
	}

	defer resp.Body.Close()

	fileExt := strings.Split(resp.Header.Get("Content-Type"), "/")[1]
	verifyImg := filepath.Join(getCwd(), "verify."+fileExt)
	fd, err := os.OpenFile(verifyImg, os.O_WRONLY, 0777)
	if err != nil {
		if os.IsNotExist(err) {
			var _err error
			fd, _err = os.Create(verifyImg)
			if _err != nil {
				panic("创建验证码文件失败：" + _err.Error())
			}
		} else {
			panic("打开验证码文件失败：" + err.Error())
		}
	}
	defer fd.Close()

	io.Copy(fd, resp.Body)        // 保存验证码文件
	openCaptchaFile(verifyImg)    // 调用外部程序打开
	captcha := readCaptchaInput() // 读取用户输入

	return captcha
}

var (
	gSession = NewSession() // 全局的 Session，调用 Init() 初始化
)

// Init 用于传入配置文件，配置全局的 Session
func Init(cfgFile string) {
	// 配置账号信息
	gSession.LoadConfig(cfgFile)

	// 登录
	gSession.Login()
}
