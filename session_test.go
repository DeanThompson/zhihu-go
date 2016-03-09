package zhihu

import (
	"testing"
)

const cfgFile = "./examples/config.json"

func Test_searchXsrf(t *testing.T) {
	s := NewSession()
	logger.Debug("_xsrf: %s", s.searchXsrf())
}

//func Test_downloadCaptcha(t *testing.T) {
//	s := NewSession("./example/config.json")
//	s.downloadCaptcha()
//}

//func Test_buildLoginForm(t *testing.T) {
//	s := &Session{}
//	s.LoadConfig()
//	values := s.buildLoginForm()
//	fmt.Println(values.Encode())
//}
