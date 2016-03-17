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

	if got, ok := user.getStringField("data-id"); ok {
		return got
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
	user.setField("data-id", dataId)
	return dataId
}

// GetBio 返回用户的 BIO
func (user *User) GetBio() string {
	if user.IsAnonymous() {
		return ""
	}

	if got, ok := user.getStringField("bio"); ok {
		return got
	}

	doc := user.Doc()

	// <span class="bio" title="程序员，用 Python 和 Go 做服务端开发。">程序员，用 Python 和 Go 做服务端开发。</span>
	bio := strip(doc.Find("span.bio").Eq(0).Text())
	user.setField("bio", bio)
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

	if got, ok := user.getStringField("gender"); ok {
		return got
	}

	doc := user.Doc()

	// <span class="item gender"><i class="icon icon-profile-male"></i></span>
	sel := doc.Find("span.gender").Find("i")
	if sel.HasClass("icon-profile-male") {
		gender = "male"
	} else {
		gender = "female"
	}
	user.setField("gender", gender)
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

// GetFollowedColumnsNum 返回用户关注的专栏数量
func (user *User) GetFollowedColumnsNum() int {
	return user.getFollowedColumnsOrTopicsNum("followed-columns-num")
}

// GetFollowedTopicsNum 返回用户关注的话题数量
func (user *User) GetFollowedTopicsNum() int {
	return user.getFollowedColumnsOrTopicsNum("followed-topics-num")
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

// GetAsks 返回用户提过的问题
func (user *User) GetAsks() []*Question {
	if user.IsAnonymous() {
		return nil
	}

	total := user.GetAsksNum()
	if total == 0 {
		return nil
	}

	page := 1
	questions := make([]*Question, 0, total)
	for page < ((total-1)/pageSize + 2) {
		link := urlJoin(user.Link, fmt.Sprintf("/asks?page=%d", page))
		doc, err := newDocumentFromUrl(link)
		if err != nil {
			return nil
		}

		doc.Find("div#zh-profile-ask-list").Children().Each(func(index int, sel *goquery.Selection) {
			a := sel.Find("a.question_link")
			title := strip(a.Text())
			href, _ := a.Attr("href")
			questionLink := makeZhihuLink(href)
			// TODO 设置关注数、回答数？
			questions = append(questions, NewQuestion(questionLink, title))
		})
		page++
	}
	return questions
}

// GetAnswers 返回用户所有的回答
func (user *User) GetAnswers() []*Answer {
	if user.IsAnonymous() {
		return nil
	}

	total := user.GetAnswersNum()
	if total == 0 {
		return nil
	}

	page := 1
	answers := make([]*Answer, 0, total)
	for page < ((total-1)/pageSize + 2) {
		link := urlJoin(user.Link, fmt.Sprintf("/answers?page=%d", page))
		doc, err := newDocumentFromUrl(link)
		if err != nil {
			return nil
		}

		doc.Find("div#zh-profile-answer-list").Children().Each(func(index int, sel *goquery.Selection) {
			a := sel.Find("a.question_link")
			qTitle := strip(a.Text())
			answerHref, _ := a.Attr("href")
			qLink := makeZhihuLink(answerHref[0:strings.Index(answerHref, "/answer")])
			question := NewQuestion(qLink, qTitle)
			thisAnswer := NewAnswer(makeZhihuLink(answerHref), question, user)

			voteText, _ := sel.Find("a.zm-item-vote-count").Attr("data-votecount")
			vote, _ := strconv.Atoi(voteText)
			thisAnswer.setUpvote(vote)

			answers = append(answers, thisAnswer)
		})
		page++
	}

	return answers
}

// GetCollections 返回用户的收藏夹
func (user *User) GetCollections() []*Collection {
	if user.IsAnonymous() {
		return nil
	}

	total := user.GetCollectionsNum()
	if total == 0 {
		return nil
	}

	page := 1
	collections := make([]*Collection, 0, total)
	for page < ((total-1)/pageSize + 2) {
		link := urlJoin(user.Link, fmt.Sprintf("/collections?page=%d", page))
		doc, err := newDocumentFromUrl(link)
		if err != nil {
			return nil
		}

		doc.Find("div.zh-profile-fav-list").Children().Each(func(index int, sel *goquery.Selection) {
			a := sel.Find("a.zm-profile-fav-item-title")
			cName := strip(a.Text())
			href, _ := a.Attr("href")
			cLink := makeZhihuLink(href)
			thisCollection := NewCollection(cLink, cName, user)
			collections = append(collections, thisCollection)
		})
	}

	return collections
}

// GetFollowedTopics 返回用户关注的话题
func (user *User) GetFollowedTopics() []string {
	if user.IsAnonymous() {
		return nil
	}

	total := user.GetFollowedTopicsNum()
	if total == 0 {
		return nil
	}

	var (
		link       = urlJoin(user.Link, "/topics")
		gotDataNum = pageSize
		offset     = 0
		topics     = make([]string, 0, total)
	)

	form := url.Values{}
	form.Set("_xsrf", user.GetXsrf())
	form.Set("start", "0")

	for gotDataNum == pageSize {
		form.Set("offset", strconv.Itoa(offset))
		body := strings.NewReader(form.Encode())
		resp, err := gSession.Ajax(link, body, link)
		if err != nil {
			logger.Error("查询关注的话题失败，用户：%s, 参数：%s", user.String(), form.Encode())
			return nil
		}
		defer resp.Body.Close()
		result := topicListResult{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			logger.Error("解析返回值 json 失败：%s", err.Error())
			return nil
		}

		topicsHtml := result.Msg[1].(string)
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(topicsHtml))
		if err != nil {
			logger.Error("解析返回的 HTML 失败：%s", err.Error())
			return nil
		}

		doc.Find("div.zm-profile-section-item").Each(func(index int, sel *goquery.Selection) {
			// TODO 定义 Topic 类，并返回 []*Topic 类型
			tName := strip(sel.Find("strong").Text())
			topics = append(topics, tName)
		})

		gotDataNum = int(result.Msg[0].(float64))
		offset += gotDataNum
	}

	return topics
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

	if got, ok := user.getStringField(cacheKey); ok {
		return got
	}

	doc := user.Doc()

	// <span class="location item" title="深圳">深圳</span>
	// <span class="business item" title="互联网">...</span>
	// <span class="education item" title="中山大学">...</span>
	value, _ := doc.Find(fmt.Sprintf("span.%s", cacheKey)).Attr("title")
	user.setField(cacheKey, value)
	return value
}

func (user *User) getFollowersNumOrFolloweesNum(cacheKey string) int {
	if user.IsAnonymous() {
		return 0
	}

	if got, ok := user.getIntField(cacheKey); ok {
		return got
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
	user.setField(cacheKey, num)
	return num
}

func (user *User) getFollowedColumnsOrTopicsNum(cacheKey string) int {
	if user.IsAnonymous() {
		return 0
	}

	if got, ok := user.getIntField(cacheKey); ok {
		return got
	}

	var selector string
	switch cacheKey {
	case "followed-topics-num":
		selector = "div.zm-profile-side-topics"
	case "followed-columns-num":
		selector = "div.zm-profile-side-columns"
	default:
		return 0
	}

	doc := user.Doc()
	result := 0
	sel := doc.Find(selector)
	if sel.Size() > 0 {
		text := sel.Parent().Find("a.zg-link-litblue").Find("strong").Text()
		result = reMatchInt(strip(text))
	}
	user.setField(cacheKey, result)
	return result
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

	if got, ok := user.getIntField(cacheKey); ok {
		return got
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
	user.setField(cacheKey, num)
	return num
}

func (user *User) getProfileNum(cacheKey string) int {
	if user.IsAnonymous() {
		return 0
	}

	if got, ok := user.getIntField(cacheKey); ok {
		return got
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
	user.setField(cacheKey, num)
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
	user.setField("bio", bio)

	// 获取关注者数量
	followersNum := reMatchInt(strip(doc.Find("div.details").Find("a").Eq(0).Text()))
	user.setField("followers-num", followersNum)

	// 获取提问数
	asksNum := reMatchInt(strip(doc.Find("div.details").Find("a").Eq(1).Text()))
	user.setField("asks-num", asksNum)

	// 获取回答数
	answersNum := reMatchInt(strip(doc.Find("div.details").Find("a").Eq(2).Text()))
	user.setField("answers-num", answersNum)

	// 获取赞同数
	agreeNum := reMatchInt(strip(doc.Find("div.details").Find("a").Eq(3).Text()))
	user.setField("agree-num", agreeNum)

	return user, nil
}
