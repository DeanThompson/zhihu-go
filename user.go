package zhihu

import "fmt"

// User 表示一个知乎用户
type User struct {
	*ZhihuPage

	userId string // 用户名
}

func NewUser(link string, userId string) *User {
	return &User{
		ZhihuPage: newZhihuPage(link),
		userId:    userId,
	}
}

// IsAnonymous 表示该用户是否匿名用户
func (user *User) IsAnonymous() bool {
	return user.userId == "匿名用户" || user.userId == "知乎用户"
}

func (user *User) String() string {
	if user.IsAnonymous() {
		return fmt.Sprintf("<User: %s>", user.userId)
	}
	return fmt.Sprintf("<User: %s - %s>", user.userId, user.Link)
}
