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
	*Page

	// userId 表示用户的知乎 ID（用户名）
	userID string
}

// NewUser 创建一个用户对象。
// link 为空的时候表示匿名用户，此时 userId 仅允许 "匿名用户" 或 "知乎用户"；
// userId 可以为空，这种情况下调用 GetUserID 会去解析用户主页
func NewUser(link string, userID string) *User {
	if link == "" && !isAnonymous(userID) {
		panic("调用 NewUser 的参数不合法")
	}

	return &User{
		Page:   newZhihuPage(link),
		userID: userID,
	}
}

// GetUserID 返回用户的知乎 ID
func (user *User) GetUserID() string {
	if user.userID != "" {
		return user.userID
	}

	doc := user.Doc()

	// <div class="title-section ellipsis">
	//   <span class="name">黄继新</span>，
	//   <span class="bio" title="和知乎在一起">和知乎在一起</span>
	// </div>
	user.userID = strip(doc.Find("div.title-section.ellipsis").Find("span.name").Text())
	return user.userID
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
	var dataID string
	btns := doc.Find("div.zm-profile-header-op-btns")
	if btns.Size() > 0 {
		// 1. 其他用户
		dataID, _ = btns.Find("button").Attr("data-id")
	} else {
		// 2. 自己
		script := doc.Find(`script[data-name="ga_vars"]`).Text()
		data := make(map[string]interface{})
		json.Unmarshal([]byte(script), &data)
		dataID = data["user_hash"].(string)
	}
	user.setField("data-id", dataID)
	return dataID
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

// GetAvatar 返回用户的头像 URL，默认的尺寸
func (user *User) GetAvatar() string {
	if user.IsAnonymous() {
		return ""
	}

	if got, ok := user.getStringField("avatar"); ok {
		return got
	}

	img := user.Doc().Find("div.body").Find("img.Avatar").First()
	avatar, _ := img.Attr("src")
	user.setField("avatar", avatar)
	return avatar
}

// GetAvatarWithSize 返回指定尺寸的的头像 URL，size 支持的值：s, xs, m, l, xl, hd, ""
func (user *User) GetAvatarWithSize(size string) string {
	defaultAvatar := user.GetAvatar()
	if defaultAvatar == "" {
		return defaultAvatar
	}

	if !validateAvatarSize(size) {
		return defaultAvatar
	}

	return replaceAvatarSize(defaultAvatar, size)
}

// GetWeiboURL 返回用户的微博主页 URL
func (user *User) GetWeiboURL() string {
	if user.IsAnonymous() {
		return ""
	}

	if got, ok := user.getStringField("weibo-url"); ok {
		return got
	}

	value := ""
	tag := user.Doc().Find("a.zm-profile-header-user-weibo")
	if tag.Size() > 0 {
		value, _ = tag.First().Attr("href")
	}
	user.setField("weibo-url", value)
	return value
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

// GetFolloweesN 返回前 n 个用户关注的人，如果 n < 0，返回所有关注的人
func (user *User) GetFolloweesN(n int) []*User {
	users, err := user.getFolloweesOrFollowers("followees", n)
	if err != nil {
		logger.Error("获取 %s 关注的人失败：%s", user.String(), err.Error())
		return nil
	}
	return users
}

// GetFollowees 返回用户关注的人
func (user *User) GetFollowees() []*User {
	return user.GetFolloweesN(-1)
}

// GetFollowersN 返回前 n 个粉丝，如果 n < 0，返回所有粉丝
func (user *User) GetFollowersN(n int) []*User {
	users, err := user.getFolloweesOrFollowers("followers", n)
	if err != nil {
		logger.Error("获取 %s 的粉丝失败：%s", user.String(), err.Error())
		return nil
	}
	return users

}

// GetFollowers 返回用户的粉丝列表
func (user *User) GetFollowers() []*User {
	return user.GetFollowersN(-1)
}

// GetAsksN 返回用户前 n 个提问，如果 n < 0, 返回所有提问
func (user *User) GetAsksN(n int) []*Question {
	if user.IsAnonymous() {
		return nil
	}

	total := user.GetAsksNum()
	if n < 0 || n > total {
		n = total
	}
	if n == 0 {
		return nil
	}

	page := 1
	questions := make([]*Question, 0, n)
	for page < ((n-1)/pageSize + 2) {
		link := urlJoin(user.Link, fmt.Sprintf("/asks?page=%d", page))
		doc, err := newDocumentFromURL(link)
		if err != nil {
			return nil
		}

		doc.Find("div#zh-profile-ask-list").Children().Each(func(index int, sel *goquery.Selection) {
			a := sel.Find("a.question_link")
			title := strip(a.Text())
			href, _ := a.Attr("href")
			questionLink := makeZhihuLink(href)
			thisQuestion := NewQuestion(questionLink, title)

			// 获取回答数
			answersNum := reMatchInt(strip(sel.Find("div.meta").Contents().Eq(4).Text()))
			thisQuestion.setAnswersNum(answersNum)

			// 获取关注数
			followersNum := reMatchInt(strip(sel.Find("div.meta").Contents().Eq(6).Text()))
			thisQuestion.setFollowersNum(followersNum)

			// 获取浏览量
			visitTimes, _ := strconv.Atoi(strip(sel.Find("div.zm-profile-vote-num").Text()))
			thisQuestion.setVisitTimes(visitTimes)

			questions = append(questions, thisQuestion)
		})

		if n > 0 && len(questions) >= n {
			return questions[:n]
		}

		page++
	}
	return questions
}

// GetAsks 返回用户所有的提问
func (user *User) GetAsks() []*Question {
	return user.GetAsksN(-1)
}

// GetAnswersN 返回用户前 n 个回答，如果 n < 0，返回所有回答
func (user *User) GetAnswersN(n int) []*Answer {
	if user.IsAnonymous() {
		return nil
	}

	total := user.GetAnswersNum()
	if n < 0 || n > total {
		n = total
	}
	if n == 0 {
		return nil
	}

	page := 1
	answers := make([]*Answer, 0, n)
	for page < ((n-1)/pageSize + 2) {
		link := urlJoin(user.Link, fmt.Sprintf("/answers?page=%d", page))
		doc, err := newDocumentFromURL(link)
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

		if n > 0 && len(answers) >= n {
			return answers[:n]
		}

		page++
	}

	return answers
}

// GetAnswers 返回用户所有的回答
func (user *User) GetAnswers() []*Answer {
	return user.GetAnswersN(-1)
}

// GetCollectionsN 返回用户前 n 个收藏夹，如果 n < 0，返回所有收藏夹
func (user *User) GetCollectionsN(n int) []*Collection {
	if user.IsAnonymous() {
		return nil
	}

	total := user.GetCollectionsNum()
	if n < 0 || n > total {
		n = total
	}
	if n == 0 {
		return nil
	}

	page := 1
	collections := make([]*Collection, 0, n)
	for page < ((n-1)/pageSize + 2) {
		link := urlJoin(user.Link, fmt.Sprintf("/collections?page=%d", page))
		doc, err := newDocumentFromURL(link)
		if err != nil {
			return nil
		}

		doc.Find("div.zm-profile-section-item").Each(func(index int, sel *goquery.Selection) {
			a := sel.Find("a.zm-profile-fav-item-title")
			cName := strip(a.Text())
			href, _ := a.Attr("href")
			cLink := makeZhihuLink(href)
			thisCollection := NewCollection(cLink, cName, user)
			collections = append(collections, thisCollection)
		})

		if n > 0 && len(collections) >= n {
			return collections[:n]
		}

		page++
	}

	return collections
}

// GetCollections 返回用户的收藏夹
func (user *User) GetCollections() []*Collection {
	return user.GetCollectionsN(-1)
}

// GetFollowedTopicsN 返回用户前 n 个关注的话题，如果 n < 0，返回所有话题
func (user *User) GetFollowedTopicsN(n int) []*Topic {
	if user.IsAnonymous() {
		return nil
	}

	total := user.GetFollowedTopicsNum()
	if n < 0 || n > total {
		n = total
	}
	if n == 0 {
		return nil
	}

	var (
		link       = urlJoin(user.Link, "/topics")
		gotDataNum = pageSize
		offset     = 0
		topics     = make([]*Topic, 0, n)
	)

	form := url.Values{}
	form.Set("_xsrf", user.GetXSRF())
	form.Set("start", "0")

	for gotDataNum == pageSize {
		form.Set("offset", strconv.Itoa(offset))
		doc, dataNum, err := newDocByNormalAjax(link, form)
		if err != nil {
			return nil
		}

		doc.Find("div.zm-profile-section-item").Each(func(index int, sel *goquery.Selection) {
			tName := strip(sel.Find("strong").Text())
			tHref, _ := sel.Find("a.zm-list-avatar-link").Attr("href")
			thisTopic := NewTopic(makeZhihuLink(tHref), tName)
			topics = append(topics, thisTopic)
		})

		if n > 0 && len(topics) >= n {
			return topics[:n]
		}

		gotDataNum = dataNum
		offset += gotDataNum
	}

	return topics
}

// GetFollowedTopics 返回用户关注的话题
func (user *User) GetFollowedTopics() []*Topic {
	return user.GetFollowedTopicsN(-1)
}

// GetLikes 返回用户赞过的回答
func (user *User) GetLikes() []*Answer {
	if user.IsAnonymous() {
		return nil
	}
	// TODO
	return nil
}

// GetVotedAnswers 是 GetLikes 的别名
func (user *User) GetVotedAnswers() []*Answer {
	return user.GetLikes()
}

// IsAnonymous 表示该用户是否匿名用户
func (user *User) IsAnonymous() bool {
	return isAnonymous(user.userID)
}

func (user *User) String() string {
	if user.IsAnonymous() {
		return fmt.Sprintf("<User: %s>", user.userID)
	}
	return fmt.Sprintf("<User: %s - %s>", user.userID, user.Link)
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

func (user *User) getFolloweesOrFollowers(eeOrEr string, limit int) ([]*User, error) {
	if user.IsAnonymous() {
		return nil, nil
	}

	if limit == 0 {
		return nil, nil
	}

	var (
		referer, ajaxURL string
		offset, totalNum int
		hashID           = user.GetDataID()
	)

	if eeOrEr == "followees" {
		referer = urlJoin(user.Link, "/followees")
		ajaxURL = makeZhihuLink("/node/ProfileFolloweesListV2")
		totalNum = user.GetFollowersNum()
	} else {
		referer = urlJoin(user.Link, "/followers")
		ajaxURL = makeZhihuLink("/node/ProfileFollowersListV2")
		totalNum = user.GetFolloweesNum()
	}

	if limit < 0 || limit > totalNum {
		limit = totalNum
	}

	form := url.Values{}
	form.Set("_xsrf", user.GetXSRF())
	form.Set("method", "next")

	users := make([]*User, 0, limit)
	for {
		form.Set("params", fmt.Sprintf(`{"offset":%d,"order_by":"created","hash_id":"%s"}`, offset, hashID))
		body := strings.NewReader(form.Encode())
		resp, err := gSession.Ajax(ajaxURL, body, referer)
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()
		result := nodeListResult{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			logger.Error("json decode failed: %s", err.Error())
			return nil, err
		}

		for _, userHTML := range result.Msg {
			thisUser, err := newUserFromHTML(userHTML)
			if err != nil {
				return nil, err
			}
			users = append(users, thisUser)
			if len(users) == limit {
				break
			}
		}

		// 已经获取了需要的数量，或者数量不够，但是已经到了最后一页
		if len(users) == limit || len(result.Msg) < pageSize {
			break
		} else {
			offset += pageSize
		}
	}
	return users, nil
}

func (user *User) setFollowersNum(value int) {
	user.setField("followers-num", value)
}

func (user *User) setAsksNum(value int) {
	user.setField("asks-num", value)
}

func (user *User) setAnswersNum(value int) {
	user.setField("answers-num", value)
}

func (user *User) setAgreeNum(value int) {
	user.setField("agree-num", value)
}

func (user *User) setBio(value string) {
	user.setField("bio", value)
}

func isAnonymous(userID string) bool {
	return userID == "匿名用户" || userID == "知乎用户"
}

func newUserFromHTML(html string) (*User, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		logger.Error("NewDocumentFromReader failed: %s", err.Error())
		return nil, err
	}

	return newUserFromSelector(doc.Selection), nil
}

func newUserFromSelector(sel *goquery.Selection) *User {
	a := sel.Find("h2.zm-list-content-title").Find("a.zg-link")
	if a.Size() == 0 {
		// 匿名用户，没有用户主页入口
		return ANONYMOUS
	}

	userId := strip(a.Text())
	link, _ := a.Attr("href")

	user := NewUser(link, userId)

	// 获取 BIO
	bio := strip(sel.Find("div.zg-big-gray").Text())
	user.setField("bio", bio)

	// 获取关注者数量
	followersNum := reMatchInt(strip(sel.Find("div.details").Find("a").Eq(0).Text()))
	user.setFollowersNum(followersNum)

	// 获取提问数
	asksNum := reMatchInt(strip(sel.Find("div.details").Find("a").Eq(1).Text()))
	user.setAsksNum(asksNum)

	// 获取回答数
	answersNum := reMatchInt(strip(sel.Find("div.details").Find("a").Eq(2).Text()))
	user.setAnswersNum(answersNum)

	// 获取赞同数
	agreeNum := reMatchInt(strip(sel.Find("div.details").Find("a").Eq(3).Text()))
	user.setAgreeNum(agreeNum)

	return user
}
