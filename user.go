package zhihu

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	ANONYMOUS = NewUser("", "匿名用户")
)

// User 表示一个知乎用户
type User struct {
	*ZhihuPage

	// userId 表示用户的知乎 ID（用户名）
	userId string
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
	if user.IsAnonymous() {
		return ""
	}

	if got, ok := user.fields["data-id"]; ok {
		return got.(string)
	}

	doc := user.Doc()

	// 分两种情况：自己和其他用户
	// 1. 其他用户
	// <div class="zm-profile-header-op-btns clearfix">
	//   <button data-follow="m:button" data-id="e22dba11081f3d71afc10b9c8c641672" class="zg-btn zg-btn-unfollow zm-rich-follow-btn">取消关注</button>
	// </div>
	//
	// 2. 自己
	// <input type="hidden" name="dest_id" value="2f5c3f612108780e7d5400d8f74ab449">
	var dataId string
	btns := doc.Find("div.zm-profile-header-op-btns")
	if btns.Size() > 0 {
		// 1. 其他用户
		dataId, _ = btns.Find("button").Attr("data-id")
	} else {
		// 2. 自己
		script := doc.Find(`script[data-name="ga_vars"]`).Text()
		data := make(map[string]interface{})
		json.Unmarshal([]byte(script), &data)
		dataId = data["user_hash"].(string)
	}
	user.fields["data-id"] = dataId
	return dataId
}

// GetBio 返回用户的 BIO
func (user *User) GetBio() string {
	if user.IsAnonymous() {
		return ""
	}

	if got, ok := user.fields["bio"]; ok {
		return got.(string)
	}

	doc := user.Doc()

	// <span class="bio" title="程序员，用 Python 和 Go 做服务端开发。">程序员，用 Python 和 Go 做服务端开发。</span>
	bio := strip(doc.Find("span.bio").Eq(0).Text())
	user.fields["bio"] = bio
	return bio
}

// GetLocation 返回用户所在地
func (user *User) GetLocation() string {
	return user.getProfile("location")
}

// GetBusiness 返回用户的所在行业
func (user *User) GetBusiness() string {
	return user.getProfile("business")
}

// GetEducation 返回用户的教育信息
func (user *User) GetEducation() string {
	return user.getProfile("education")
}

// GetGender 返回用户的性别（male/female/unknown）
func (user *User) GetGender() string {
	gender := "unknown"
	if user.IsAnonymous() {
		return gender
	}

	if got, ok := user.fields["gender"]; ok {
		return got.(string)
	}

	doc := user.Doc()

	// <span class="item gender"><i class="icon icon-profile-male"></i></span>
	sel := doc.Find("span.gender").Find("i")
	if sel.HasClass("icon-profile-male") {
		gender = "male"
	} else {
		gender = "female"
	}
	user.fields["gender"] = gender
	return gender
}

// GetFollowersNum 返回用户的粉丝数量
func (user *User) GetFollowersNum() int {
	return user.getFollowersNumOrFolloweesNum("followers-num")
}

// GetFolloweesNum 返回用户关注的数量
func (user *User) GetFolloweesNum() int {
	return user.getFollowersNumOrFolloweesNum("followees-num")
}

// GetAgreeNum 返回用户的点赞数
func (user *User) GetAgreeNum() int {
	return user.getAgreeOrThanksNum("agree-num")
}

// GetThanksNum 返回用户的感谢数
func (user *User) GetThanksNum() int {
	return user.getAgreeOrThanksNum("thanks-num")
}

// GetAsksNum 返回用户的提问数
func (user *User) GetAsksNum() int {
	return user.getProfileNum("asks-num")
}

// GetAnswersNum 返回用户的回答数
func (user *User) GetAnswersNum() int {
	return user.getProfileNum("answers-num")
}

// GetPostsNum 返回用户的专栏文章数量
func (user *User) GetPostsNum() int {
	return user.getProfileNum("posts-num")
}

// GetCollectionsNum 返回用户的收藏夹数量
func (user *User) GetCollectionsNum() int {
	return user.getProfileNum("collections-num")
}

// GetLogsNum 返回用户公共编辑数量
func (user *User) GetLogsNum() int {
	return user.getProfileNum("logs-num")
}

// GetFollowees 返回用户关注的人
func (user *User) GetFollowees() []*User {
	users, err := user.getFolloweesOrFollowers("followees")
	if err != nil {
		logger.Error("获取 %s 关注的人失败：%s", user.String(), err.Error())
		return nil
	}
	return users
}

// GetFollowers 返回用户的粉丝列表
func (user *User) GetFollowers() []*User {
	users, err := user.getFolloweesOrFollowers("followers")
	if err != nil {
		logger.Error("获取 %s 的粉丝失败：%s", user.String(), err.Error())
		return nil
	}
	return users
}

// TODO GetAsks 返回用户提过的问题
func (user *User) GetAsks() []*Question {
	if user.IsAnonymous() {
		return nil
	}
	return nil
}

// TODO GetAnswers 返回用户所有的回答
func (user *User) GetAnswers() []*Answer {
	if user.IsAnonymous() {
		return nil
	}

	return nil
}

// TODO GetCollections 返回用户的收藏夹
func (user *User) GetCollections() []*Collection {
	if user.IsAnonymous() {
		return nil
	}

	return nil
}

// TODO GetLikes 返回用户赞过的回答
func (user *User) GetLikes() []*Answer {
	if user.IsAnonymous() {
		return nil
	}

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

func (user *User) getProfile(cacheKey string) string {
	if user.IsAnonymous() {
		return ""
	}

	if got, ok := user.fields[cacheKey]; ok {
		return got.(string)
	}

	doc := user.Doc()

	// <span class="location item" title="深圳">深圳</span>
	// <span class="business item" title="互联网">...</span>
	// <span class="education item" title="中山大学">...</span>
	value, _ := doc.Find(fmt.Sprintf("span.%s", cacheKey)).Attr("title")
	user.fields[cacheKey] = value
	return value
}

func (user *User) getFollowersNumOrFolloweesNum(cacheKey string) int {
	if user.IsAnonymous() {
		return 0
	}

	if got, ok := user.fields[cacheKey]; ok {
		return got.(int)
	}

	var index int
	switch cacheKey {
	case "followees-num":
		index = 0
	case "followers-num":
		index = 1
	default:
		return 0
	}

	doc := user.Doc()

	// <div class="zm-profile-side-following zg-clear">
	//   <a class="item" href="/people/jixin/followees">
	//     <span class="zg-gray-normal">关注了</span><br><strong>9190</strong><label> 人</label>
	//   </a>
	//   <a class="item" href="/people/jixin/followers">
	//     <span class="zg-gray-normal">关注者</span><br><strong>754769</strong><label> 人</label>
	//   </a>
	// </div>
	value := doc.Find("div.zm-profile-side-following a strong").Eq(index).Text()
	num, _ := strconv.Atoi(value)
	user.fields[cacheKey] = num
	return num
}

func (user *User) getAgreeOrThanksNum(cacheKey string) int {
	if user.IsAnonymous() {
		return 0
	}

	var selector string
	switch cacheKey {
	case "agree-num":
		selector = "span.zm-profile-header-user-agree > strong"
	case "thanks-num":
		selector = "span.zm-profile-header-user-thanks > strong"
	default:
		return 0
	}

	if got, ok := user.fields[cacheKey]; ok {
		return got.(int)
	}

	doc := user.Doc()

	// <div class="zm-profile-header-operation zg-clear ">
	//   <div class="zm-profile-header-info-list">
	//     <span class="zm-profile-header-info-title">获得</span>
	//     <span class="zm-profile-header-user-agree"><span class="zm-profile-header-icon"></span><strong>68200</strong>赞同</span>
	//     <span class="zm-profile-header-user-thanks"><span class="zm-profile-header-icon"></span><strong>17511</strong>感谢</span>
	//   </div>
	// </div>
	num, _ := strconv.Atoi(doc.Find(selector).Text())
	user.fields[cacheKey] = num
	return num
}

func (user *User) getProfileNum(cacheKey string) int {
	if user.IsAnonymous() {
		return 0
	}

	if got, ok := user.fields[cacheKey]; ok {
		return got.(int)
	}

	var index int
	switch cacheKey {
	case "asks-num":
		index = 0
	case "answers-num":
		index = 1
	case "posts-num":
		index = 2
	case "collections-num":
		index = 3
	case "logs-num":
		index = 4
	default:
		return 0
	}

	doc := user.Doc()

	// <div class="profile-navbar clearfix">
	//   <a class="item home first active" href="/people/jixin"><i class="icon icon-profile-tab-home"></i><span class="hide-text">主页</span></a>
	//   <a class="item " href="/people/jixin/asks"> 提问 <span class="num">1336</span></a>
	//   <a class="item " href="/people/jixin/answers"> 回答 <span class="num">785</span></a>
	//   <a class="item " href="/people/jixin/posts"> 专栏文章 <span class="num">91</span></a>
	//   <a class="item " href="/people/jixin/collections"> 收藏 <span class="num">44</span></a>
	//   <a class="item " href="/people/jixin/logs"> 公共编辑 <span class="num">51471</span></a>
	// </div>
	value := doc.Find("div.profile-navbar").Find("span.num").Eq(index).Text()
	num, _ := strconv.Atoi(value)
	user.fields[cacheKey] = num
	return num
}

func (user *User) getFolloweesOrFollowers(eeOrEr string) ([]*User, error) {
	if user.IsAnonymous() {
		return nil, nil
	}

	var (
		referer, ajaxUrl string
		offset, totalNum int
		hashId           = user.GetDataID()
	)

	if eeOrEr == "followees" {
		referer = urlJoin(user.Link, "/followees")
		ajaxUrl = makeZhihuLink("/node/ProfileFolloweesListV2")
		totalNum = user.GetFollowersNum()
	} else {
		referer = urlJoin(user.Link, "/followers")
		ajaxUrl = makeZhihuLink("/node/ProfileFollowersListV2")
		totalNum = user.GetFolloweesNum()
	}

	form := url.Values{}
	form.Set("_xsrf", user.GetXsrf())
	form.Set("method", "next")

	users := make([]*User, 0, totalNum)
	for {
		form.Set("params", fmt.Sprintf(`{"offset":%d,"order_by":"created","hash_id":"%s"}`, offset, hashId))
		body := strings.NewReader(form.Encode())
		resp, err := gSession.Ajax(ajaxUrl, body, referer)
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()
		result := dataListResult{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			logger.Error("json decode failed: %s", err.Error())
			return nil, err
		}

		for _, userHtml := range result.Msg {
			thisUser, err := newUserFromHTML(userHtml)
			if err != nil {
				return nil, err
			}
			users = append(users, thisUser)
		}

		if len(result.Msg) < pageSize {
			break
		} else {
			offset += pageSize
		}
	}
	return users, nil
}

func (user *User) setStringAttr(attr, value string) {
	user.fields[attr] = value
}

func (user *User) setIntAttr(attr string, value int) {
	user.fields[attr] = value
}

func isAnonymous(userId string) bool {
	return userId == "匿名用户" || userId == "知乎用户"
}

func newUserFromHTML(html string) (*User, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		logger.Error("NewDocumentFromReader failed: %s", err.Error())
		return nil, err
	}

	a := doc.Find("h2.zm-list-content-title").Find("a.zg-link")
	userId := strip(a.Text())
	link, _ := a.Attr("href")

	user := NewUser(link, userId)

	// 获取 BIO
	bio := strip(doc.Find("div.zg-big-gray").Text())
	user.setStringAttr("bio", bio)

	// 获取关注者数量
	followersNum := reMatchInt(strip(doc.Find("div.details").Find("a").Eq(0).Text()))
	user.setIntAttr("followers-num", followersNum)

	// 获取提问数
	asksNum := reMatchInt(strip(doc.Find("div.details").Find("a").Eq(1).Text()))
	user.setIntAttr("asks-num", asksNum)

	// 获取回答数
	answersNum := reMatchInt(strip(doc.Find("div.details").Find("a").Eq(2).Text()))
	user.setIntAttr("answers-num", answersNum)

	// 获取赞同数
	agreeNum := reMatchInt(strip(doc.Find("div.details").Find("a").Eq(3).Text()))
	user.setIntAttr("agree-num", agreeNum)

	return user, nil
}
