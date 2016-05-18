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
	*Page

	// creator 是该收藏夹的创建者
	creator *User

	// name 是该收藏夹的名称
	name string
}

// NewCollection 创建一个收藏夹对象，返回 *Collection
func NewCollection(link string, name string, creator *User) *Collection {
	if !validCollectionURL(link) {
		panic("收藏夹链接不正确：" + link)
	}

	return &Collection{
		Page:    newZhihuPage(link),
		creator: creator,
		name:    name,
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

// GetFollowersN 返回 n 个关注该收藏夹的用户，如果 n < 0，返回所有关注者
func (c *Collection) GetFollowersN(n int) []*User {
	var (
		link = urlJoin(c.Link, "/followers")
		xsrf = c.GetXSRF()
	)
	users, err := ajaxGetFollowers(link, xsrf, n)
	if err != nil {
		return nil
	}
	return users
}

// GetFollowers 返回关注该收藏夹的用户
func (c *Collection) GetFollowers() []*User {
	return c.GetFollowersN(c.GetFollowersNum())
}

// GetQuestionsN 返回前 n 个问题，如果 n < 0，返回所有问题
func (c *Collection) GetQuestionsN(n int) []*Question {
	if n == 0 {
		return nil
	}

	// 先获取第一页的问题
	questions := getQuestionsFromDoc(c.Doc())

	totalPages := c.totalPages()
	if totalPages == 1 {
		if n < 0 || n > len(questions) {
			return questions
		}
		return questions[0:n]
	}

	// 再分页查询其他问题
	currentPage := 2
	for currentPage <= totalPages {
		link := fmt.Sprintf("%s?page=%d", c.Link, currentPage)
		doc, err := newDocumentFromURL(link)
		if err != nil {
			logger.Error("解析页面失败：%s, %s", link, err.Error())
			return nil
		}

		newQuestions := getQuestionsFromDoc(doc)
		questions = append(questions, newQuestions...)
		if n > 0 && len(questions) >= n {
			return questions[0:n]
		}
		currentPage++
	}

	return questions
}

// GetQuestions 返回收藏夹里所有的问题
func (c *Collection) GetQuestions() []*Question {
	return c.GetQuestionsN(-1)
}

// GetAnswersN 返回 n 个回答，如果 n < 0，返回所有回答
func (c *Collection) GetAnswersN(n int) []*Answer {
	if n == 0 {
		return nil
	}

	// 先获取第一页的回答
	answers := getAnswersFromDoc(c.Doc())

	totalPages := c.totalPages()
	if totalPages == 1 {
		if n < 0 || n > len(answers) {
			return answers
		}
		return answers[0:n]
	}

	// 在分页查询
	currentPage := 2
	for currentPage <= totalPages {
		link := fmt.Sprintf("%s?page=%d", c.Link, currentPage)
		doc, err := newDocumentFromURL(link)
		if err != nil {
			logger.Error("解析页面失败：%s, %s", link, err.Error())
			return nil
		}

		newAnswers := getAnswersFromDoc(doc)
		answers = append(answers, newAnswers...)
		if n > 0 && len(answers) >= n {
			return answers[0:n]
		}
		currentPage++
	}
	return answers
}

// GetAnswers 返回收藏夹里所有的回答
func (c *Collection) GetAnswers() []*Answer {
	return c.GetAnswersN(-1)
}

// GetQuestionsNum 返回收藏夹的问题数量
func (c *Collection) GetQuestionsNum() int {
	if value, ok := c.getIntField("question-num"); ok {
		return value
	}

	// 根据分页情况来计算问题数量
	// 收藏夹页面，每一页固定 10 个问题，每个问题下可能有多个答案；
	totalPages := c.totalPages()
	lastPage := c.Doc()

	if totalPages > 1 {
		lp, err := newDocumentFromURL(fmt.Sprintf("%s?page=%d", c.Link, totalPages))
		if err != nil {
			logger.Error("获取收藏夹最后一页失败：%s", err.Error())
			return 0
		}
		lastPage = lp
	}

	numOnLastPage := lastPage.Find("#zh-list-answer-wrap h2.zm-item-title").Size()
	rv := (totalPages-1)*10 + numOnLastPage
	c.setField("question-num", rv)
	return rv
}

// GetAnswersNum 返回收藏夹的答案数量
// 获取答案数量有这几种方式：
// 	1. 在收藏夹页面（/collections/1234567），遍历每一页，累计每页的回答数量。总请求数等于分页数。
//	2. 在收藏夹创建者的个人主页，收藏夹栏目（people/xxyy/collections），有每个收藏夹的简介，
//     其中就有回答数。遍历每一页（20个/页），找到对应的收藏夹，然后获取回答数。
//     总请求数不确定，最好情况下 1 次；但考虑到每个用户的收藏夹并不会很多（如达到100个），可以认为最坏情况下需要 5 次。
// 最终的方案可以综合以上两种方式，以收藏夹页面分页数做依据：
//  如果页数大于 3（经验值），则采用方法 2；否则用方法 1
// 希望能通过这样的方式来减少请求数，获得更好的性能。
func (c *Collection) GetAnswersNum() int {
	if value, ok := c.getIntField("answer-num"); ok {
		return value
	}

	rv := 0
	totalPages := c.totalPages()
	if totalPages > 3 {
		// 从个人主页上获取
		page := 1
		linkFmt := urlJoin(c.GetCreator().Link, "/collections?page=%d")
		collectionHref := strings.Split(c.Link, "zhihu.com")[1]
		selector := fmt.Sprintf(`a.zm-profile-fav-item-title[href="%s"]`, collectionHref)
		for {
			creatorCollectionLink := fmt.Sprintf(linkFmt, page)
			doc, err := newDocumentFromURL(creatorCollectionLink)
			if err != nil {
				logger.Error("获取用户的收藏夹主页失败：%s", err.Error())
				return 0
			}
			titleTag := doc.Find(selector).First()
			if titleTag.Size() == 1 {
				rv = reMatchInt(titleTag.Parent().Next().Contents().Eq(0).Text())
				break
			} else {
				// 本页没找到，下一页
				if doc.Find("div.border-pager").Size() == 0 {
					return 0
				} else {
					pages := getTotalPages(doc)
					if page == pages {
						return 0
					}
					page++
				}
			}
		}
	} else {
		selector := "#zh-list-answer-wrap div.zm-item-fav"
		rv = c.Doc().Find(selector).Size()
		currentPage := 2
		for currentPage <= totalPages {
			link := fmt.Sprintf("%s?page=%d", c.Link, currentPage)
			doc, err := newDocumentFromURL(link)
			if err != nil {
				logger.Error("解析页面失败：%s, %s", link, err.Error())
				return 0
			}
			rv += doc.Find(selector).Size()
			currentPage++
		}
	}
	c.setField("answer-num", rv)
	return rv
}

// GetCommentsNum 返回评论数量
func (c *Collection) GetCommentsNum() int {
	if value, ok := c.getIntField("comment-num"); ok {
		return value
	}

	doc := c.Doc()
	text := strip(doc.Find("div#zh-list-meta-wrap  a.toggle-comment").Text())
	rv := reMatchInt(text)
	c.setField("comment-num", rv)
	return rv
}

func (c *Collection) String() string {
	return fmt.Sprintf("<Collection: %s - %s>", c.GetName(), c.Link)
}

func ajaxGetFollowers(link string, xsrf string, total int) ([]*User, error) {
	if total == 0 {
		return nil, nil
	}

	var (
		offset     = 0
		gotDataNum = pageSize
		initCap    = total
	)

	if initCap < 0 {
		initCap = pageSize
	}
	users := make([]*User, 0, initCap)

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

		if total > 0 && len(users) >= total {
			return users[:total], nil
		}

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
	var answers []*Answer
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
