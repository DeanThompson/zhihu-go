package zhihu

import "fmt"

// User 表示一个知乎用户
type User struct {
	*ZhihuPage

	// userId 表示用户的知乎 ID（用户名）
	userId string

	// fields 是字段缓存，避免重复解析页面
	fields map[string]interface{}
}

// NewUser 创建一个用户对象
// link 为空的时候表示匿名用户，此时 userId 仅允许 "匿名用户" 或 "知乎用户"
// userId 可以为空，这种情况下调用 GetUserID 会去解析用户主页
func NewUser(link string, userId string) *User {
	if link == "" && !isAnonymous(userId) {
		panic("调用 NewUser 的参数不合法")
	}

	return &User{
		ZhihuPage: newZhihuPage(link),
		userId:    userId,
		fields:    make(map[string]interface{}),
	}
}

// GetUserID 返回用户的知乎 ID
func (user *User) GetUserID() string {
	if user.userId != "" {
		return user.userId
	}

	doc := user.Doc()

	// <div class="title-section ellipsis">
	//   <span class="name">黄继新</span>，
	//   <span class="bio" title="和知乎在一起">和知乎在一起</span>
	// </div>
	user.userId = strip(doc.Find("div.title-section.ellipsis").Find("span.name").Text())
	return user.userId
}

// GetDataID 返回用户的 data-id
func (user *User) GetDataID() string {
	if got, ok := user.fields["data-id"]; ok {
		return got.(string)
	}

	doc := user.Doc()

	// <button data-follow="m:button" data-id="b6f80220378c8b0b78175dd6a0b9c680" class="zg-btn zg-btn-unfollow zm-rich-follow-btn">
	//   取消关注
	// </button>
	dataId, _ := doc.Find("button.zg-btn.zm-rich-follow-btn").Attr("data-id")
	user.fields["data-id"] = dataId
	return dataId
}

// GetGender 返回用户的性别
func (user *User) GetGender() string {
	if got, ok := user.fields["gender"]; ok {
		return got.(string)
	}

	doc := user.Doc()

	gender := "unknown"

	// <span class="item gender">
	// 	<i class="icon icon-profile-male"></i>
	// </span>
	sel := doc.Find("span.gender").Find("i")
	if sel.HasClass("icon-profile-male") {
		gender = "male"
	} else {
		gender = "female"
	}
	user.fields["gender"] = gender
	return gender
}

// TODO GetFollowersNum 返回用户的粉丝数量
func (user *User) GetFollowersNum() int {
	return 0
}

// TODO GetFolloweesNum 返回用户关注的数量
func (user *User) GetFolloweesNum() int {
	return 0
}

// TODO GetAgreeNum 返回用户的点赞数
func (user *User) GetAgreeNum() int {
	return 0
}

// TODO GetThanksNum 返回用户的感谢数
func (user *User) GetThanksNum() int {
	return 0
}

// TODO GetAsksNum 返回用户的提问数
func (user *User) GetAsksNum() int {
	return 0
}

// TODO GetAnswersNum 返回用户的回答数
func (user *User) GetAnswersNum() int {
	return 0
}

// TODO GetCollectionsNum 返回用户的收藏夹数量
func (user *User) GetCollectionsNum() int {
	return 0
}

// TODO GetFollowees 返回用户关注的人
func (user *User) GetFollowees() []*User {
	return nil
}

// TODO GetFollowers 返回用户的粉丝列表
func (user *User) GetFollowers() []*User {
	return nil
}

// TODO GetAsks 返回用户提过的问题
func (user *User) GetAsks() []*Question {
	return nil
}

// TODO GetAnswers 返回用户所有的回答
func (user *User) GetAnswers() []*Answer {
	return nil
}

// TODO GetCollections 返回用户的收藏夹
func (user *User) GetCollections() []*Collection {
	return nil
}

// TODO GetLikes 返回用户赞过的回答
func (user *User) GetLikes() []*Answer {
	return nil
}

// GetVotedAnswers 是 GetLikes 的别名
func (user *User) GetVotedAnswers() []*Answer {
	return user.GetLikes()
}

// IsAnonymous 表示该用户是否匿名用户
func (user *User) IsAnonymous() bool {
	return isAnonymous(user.userId)
}

func (user *User) String() string {
	if user.IsAnonymous() {
		return fmt.Sprintf("<User: %s>", user.userId)
	}
	return fmt.Sprintf("<User: %s - %s>", user.userId, user.Link)
}

func isAnonymous(userId string) bool {
	return userId == "匿名用户" || userId == "知乎用户"
}
