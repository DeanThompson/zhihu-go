package zhihu

import (
	"fmt"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

// Question 表示一个知乎问题，可以用于获取其标题、详情、答案等信息
type Question struct {
	*ZhihuPage

	// title 是该问题的标题
	title string

	// fields 是字段缓存，避免重复解析页面
	fields map[string]interface{}
}

// NewQuestion 通过给定的 URL 创建一个 Question 对象
func NewQuestion(link string, title string) *Question {
	if !validQuestionURL(link) {
		panic("问题链接不正确: " + link)
	}

	return &Question{
		ZhihuPage: newZhihuPage(link),
		title:     title,
		fields:    make(map[string]interface{}),
	}
}

// GetTitle 获取问题标题
func (q *Question) GetTitle() string {
	if q.title != "" {
		return q.title
	}

	doc := q.Doc()
	q.title = strip(doc.Find("h2.zm-item-title").First().Text())
	return q.title
}

// GetDetail 获取问题描述
func (q *Question) GetDetail() string {
	if got, ok := q.fields["detail"]; ok {
		return got.(string)
	}

	doc := q.Doc()
	detail := strip(doc.Find("div#zh-question-detail").First().Text())
	q.fields["detail"] = detail
	return detail
}

// GetAnswerNum 获取问题回答数量
func (q *Question) GetAnswersNum() int {
	if got, ok := q.fields["answer-num"]; ok {
		return got.(int)
	}

	doc := q.Doc()
	data, exists := doc.Find("h3#zh-question-answer-num").Attr("data-num")
	answerNum := 0
	if exists {
		answerNum, _ = strconv.Atoi(data)
	}
	q.fields["answer-num"] = answerNum
	return answerNum
}

// GetFollowersNum 获取问题关注数量
func (q *Question) GetFollowersNum() int {
	if got, ok := q.fields["followers-num"]; ok {
		return got.(int)
	}

	doc := q.Doc()
	text := doc.Find("div.zg-gray-normal>a>strong").Text()
	followersNum, _ := strconv.Atoi(text)
	q.fields["followers-num"] = followersNum
	return followersNum
}

// GetTopics 获取问题的话题列表
func (q *Question) GetTopics() []string {
	if got, ok := q.fields["topics"]; ok {
		return got.([]string)
	}

	doc := q.Doc()
	selection := doc.Find("a.zm-item-tag")
	topics := selection.Map(func(index int, node *goquery.Selection) string {
		return strip(node.Text())
	})
	q.fields["topics"] = topics
	return topics
}

// TODO GetAllAnswers 获取问题的所有答案
func (q *Question) GetAllAnswers() []*Answer {
	// 1. 首页的回答
	// 2. "更多"，调用 Ajax 接口
	return nil
}

// TODO GetTopXAnswer 获取问题 Top X 的答案
func (q *Question) GetTopXAnswer(x int) []*Answer {
	return nil
}

// TODO GetTopAnswer 获取问题排名第一的答案
func (q *Question) GetTopAnswer() *Answer {
	answers := q.GetTopXAnswer(1)
	if len(answers) >= 1 {
		return answers[0]
	}
	return nil
}

// GetVisitTimes 获取问题的访问次数
func (q *Question) GetVisitTimes() int {
	if got, ok := q.fields["visit-times"]; ok {
		return got.(int)
	}

	doc := q.Doc()
	content, exists := doc.Find(`meta[itemprop="visitsCount"]`).Attr("content")
	visitTimes := 0
	if exists {
		visitTimes, _ = strconv.Atoi(content)
	}
	q.fields["visit-times"] = visitTimes
	return visitTimes
}

func (q *Question) String() string {
	return fmt.Sprintf("<Question: %s - %s>", q.GetTitle(), q.Link)
}

// getAnswersOnIndex 解析问题页面，返回第一页的回答
func (q *Question) getAnswersOnIndex() []*Answer {
	var (
		answerPerPage = 20
		totalNum      = q.GetAnswersNum()
	)
	answers := make([]*Answer, minInt(answerPerPage, totalNum))

	doc := q.Doc()

	doc.Find("div.zm-item-answer").Each(func(index int, sel *goquery.Selection) {
		answers = append(answers, q.processSingleAnswer(sel))
	})
	return nil
}

// TODO getMoreAnswers 处理 “更多” 回答，调用 Ajax 接口
func (q *Question) getMoreAnswers() []*Answer {
	return nil
}

// TODO processSingleAnswer 处理一个回答的 HTML 片段，
// 这段 HTML 可能来自问题页面，也可能来自 Ajax 接口
func (q *Question) processSingleAnswer(sel *goquery.Selection) *Answer {
	// 1. 获取链接
	answerHref, _ := sel.Find("a.answer-date-link").Attr("href")
	answerLink := makeZhihuLink(answerHref)

	// 2. 获取作者
	authorSel := sel.Find("div.zm-item-answer-author-info")
	var author *User
	if !authorSel.Has("a.author-link") {
		// 匿名用户
		author = NewUser("", "匿名用户")
	} else {
		// 具名用户
		x := authorSel.Find("a.author-link")
		userId := strip(x.Text())
		userHref, _ := x.Attr("href")
		author = NewUser(makeZhihuLink(userHref), userId)
	}

	answer := NewAnswer(answerLink, q, author)

	// 3. 获取赞同数
	dataIsOwner, _ := sel.Attr("data-isowner")
	isOwner := dataIsOwner == "1" // 判断是否本人的回答
	var voteText string
	if isOwner {
		voteText = strip(sel.Find("a.zm-item-vote-count").Text())
	} else {
		voteText = strip(sel.Find("div.zm-votebar").Find("span.count").Text())
	}
	voteCount, _ := strconv.Atoi(voteText)
	answer.setUpvote(voteCount)

	// TODO 4. 获取内容
	return answer
}
