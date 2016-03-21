package zhihu

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Collection 是一个知乎的收藏夹页面
type Collection struct {
	*ZhihuPage

	// creator 是该收藏夹的创建者
	creator *User

	// name 是该收藏夹的名称
	name string
}

func NewCollection(link string, name string, creator *User) *Collection {
	if !validCollectionURL(link) {
		panic("收藏夹链接不正确：" + link)
	}

	return &Collection{
		ZhihuPage: newZhihuPage(link),
		creator:   creator,
		name:      name,
	}
}

// GetName 返回收藏夹的名字
func (c *Collection) GetName() string {
	if c.name != "" {
		return c.name
	}

	doc := c.Doc()

	// <h2 class="zm-item-title zm-editable-content" id="zh-fav-head-title">
	//   恩恩恩 大力一点，不要停～
	// </h2>
	c.name = strip(doc.Find("h2#zh-fav-head-title").Text())
	return c.name
}

// GetCreator 返回收藏夹的创建者
func (c *Collection) GetCreator() *User {
	if c.creator != nil {
		return c.creator
	}

	doc := c.Doc()

	// <h2 class="zm-list-content-title">
	//   <a href="/people/leonyoung">李阳良</a>
	// </h2>
	sel := doc.Find("h2.zm-list-content-title a")
	userId := strip(sel.Text())
	linkPath, _ := sel.Attr("href")
	c.creator = NewUser(makeZhihuLink(linkPath), userId)
	return c.creator
}

// GetFollowersNum 返回收藏夹的关注者数量
func (c *Collection) GetFollowersNum() int {
	if got, ok := c.getIntField("followers-num"); ok {
		return got
	}

	doc := c.Doc()

	// <a href="/collection/19653044/followers" data-za-c="collection" ,="" data-za-a="visit_collection_followers" data-za-l="collection_followers_count">
	//   7516
	// </a>
	text := strip(doc.Find(`a[data-za-a="visit_collection_followers"]`).Text())
	num, _ := strconv.Atoi(text)
	c.setField("followers-num", num)
	return num
}

// GetFollowers 返回关注该收藏夹的用户
func (c *Collection) GetFollowers() []*User {
	var (
		link  = urlJoin(c.Link, "/followers")
		xsrf  = c.GetXsrf()
		total = c.GetFollowersNum()
	)
	users, err := ajaxGetFollowers(link, xsrf, total)
	if err != nil {
		return nil
	}
	return users
}

// GetQuestions 返回收藏夹里所有的问题
func (c *Collection) GetQuestions() []*Question {
	// 先获取第一页的问题
	questions := getQuestionsFromDoc(c.Doc())

	totalPages := c.totalPages()
	if totalPages == 1 {
		return questions
	}

	// 再分页查询其他问题
	currentPage := 2
	for currentPage <= totalPages {
		link := fmt.Sprintf("%s?page=%d", c.Link, currentPage)
		doc, err := newDocumentFromUrl(link)
		if err != nil {
			logger.Error("解析页面失败：%s, %s", link, err.Error())
			return nil
		}

		newQuestions := getQuestionsFromDoc(doc)
		questions = append(questions, newQuestions...)
		currentPage++
	}

	return questions
}

// GetAllAnswers 返回收藏夹里所有的回答
func (c *Collection) GetAllAnswers() []*Answer {
	// 先获取第一页的回答
	answers := getAnswersFromDoc(c.Doc())

	totalPages := c.totalPages()
	if totalPages == 1 {
		return answers
	}

	// 在分页查询
	currentPage := 2
	for currentPage <= totalPages {
		link := fmt.Sprintf("%s?page=%d", c.Link, currentPage)
		doc, err := newDocumentFromUrl(link)
		if err != nil {
			logger.Error("解析页面失败：%s, %s", link, err.Error())
			return nil
		}

		newAnswers := getAnswersFromDoc(doc)
		answers = append(answers, newAnswers...)
		currentPage++
	}
	return answers
}

func (c *Collection) String() string {
	return fmt.Sprintf("<Collection: %s - %s>", c.GetName(), c.Link)
}

func (c *Collection) totalPages() int {
	pager := c.Doc().Find("div.zm-invite-pager")
	if pager.Size() == 0 {
		// 只有一页
		return 1
	}
	text := pager.Find("span").Eq(-2).Text()
	pages, _ := strconv.Atoi(text)
	return pages
}

func ajaxGetFollowers(link string, xsrf string, total int) ([]*User, error) {
	if total < 0 {
		total = 0
	}

	var (
		offset     = 0
		gotDataNum = pageSize
		users      = make([]*User, 0, total)
	)

	form := url.Values{}
	form.Set("_xsrf", xsrf)

	for gotDataNum == pageSize {
		form.Set("offset", strconv.Itoa(offset))
		doc, dataNum, err := newDocByNormalAjax(link, form)
		if err != nil {
			return nil, err
		}

		doc.Find("div.zm-profile-card").Each(func(index int, sel *goquery.Selection) {
			thisUser := newUserFromSelector(sel)
			users = append(users, thisUser)
		})

		gotDataNum = dataNum
		offset += gotDataNum
	}
	return users, nil
}

func newDocByNormalAjax(link string, form url.Values) (*goquery.Document, int, error) {
	gotDataNum := 0
	body := strings.NewReader(form.Encode())
	resp, err := gSession.Ajax(link, body, link)
	if err != nil {
		logger.Error("查询关注的话题失败, 链接：%s, 参数：%s，错误：%s", link, form.Encode(), err.Error())
		return nil, gotDataNum, err
	}

	defer resp.Body.Close()
	result := normalAjaxResult{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		logger.Error("解析返回值 json 失败：%s", err.Error())
		return nil, gotDataNum, err
	}

	topicsHtml := result.Msg[1].(string)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(topicsHtml))
	if err != nil {
		logger.Error("解析返回的 HTML 失败：%s", err.Error())
		return nil, gotDataNum, err
	}
	gotDataNum = int(result.Msg[0].(float64))
	return doc, gotDataNum, err
}

func getQuestionsFromDoc(doc *goquery.Document) []*Question {
	questions := make([]*Question, 0, pageSize)
	items := doc.Find("div#zh-list-answer-wrap").Find("h2.zm-item-title")
	items.Each(func(index int, sel *goquery.Selection) {
		a := sel.Find("a")
		qTitle := strip(a.Text())
		qHref, _ := a.Attr("href")
		thisQuestion := NewQuestion(makeZhihuLink(qHref), qTitle)
		questions = append(questions, thisQuestion)
	})
	return questions
}

func getAnswersFromDoc(doc *goquery.Document) []*Answer {
	answers := make([]*Answer, 0)
	var lastQuestion *Question

	doc.Find("div.zm-item").Each(func(index int, sel *goquery.Selection) {
		// 回答
		contentTag := sel.Find("div.zm-item-rich-text")
		if contentTag.Size() == 0 {
			// 回答被建议修改
			reason := strip(sel.Find("div.answer-status").Text())
			logger.Warn("忽略一个问题，原因：%s", reason)
			return
		}

		// 获取问题，如果同一个问题下收藏了多个回答，则除了第一个外，后面的回答的 HTML 部分，
		// 也就是 div.zm-item 里面不会有该问题的链接（a 标签），所以用 lastQuestion 标记
		// 最近的一个问题
		var thisQuestion *Question
		if qTag := sel.Find("h2.zm-item-title").Find("a"); qTag.Size() > 0 {
			qTitle := strip(qTag.Text())
			qHref, _ := qTag.Attr("href")
			thisQuestion = NewQuestion(makeZhihuLink(qHref), qTitle)
			lastQuestion = thisQuestion
		} else {
			thisQuestion = lastQuestion
		}

		// 答主
		author := newUserFromAnswerAuthorTag(sel.Find("div.zm-item-answer-author-info"))

		answerHref, _ := contentTag.Attr("data-entry-url")
		voteText, _ := sel.Find("a.zm-item-vote-count").Attr("data-votecount")
		vote, _ := strconv.Atoi(voteText)
		thisAnswer := NewAnswer(makeZhihuLink(answerHref), thisQuestion, author)
		thisAnswer.setUpvote(vote)

		answers = append(answers, thisAnswer)
	})
	return answers
}
