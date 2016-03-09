package zhihu

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Answer struct {
	Link string

	// HTML document
	doc *goquery.Document

	question *Question

	author *User

	fields map[string]interface{}
}

func NewAnswer(link string, question *Question, author *User) *Answer {
	return &Answer{
		Link:     link,
		question: question,
		author:   author,
		fields:   make(map[string]interface{}),
	}
}

// Doc 用于获取当前问题页面的 HTML document，惰性求值
func (a *Answer) Doc() *goquery.Document {
	if a.doc != nil {
		return a.doc
	}

	resp, err := gSession.Get(a.Link)
	if err != nil {
		logger.Error("查询答案页面失败：%s", err.Error())
		return nil
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		logger.Error("解析页面失败：%s", err.Error())
		return nil
	}

	a.doc = doc
	return a.doc
}

func (a *Answer) GetQuestion() *Question {
	if a.question != nil {
		return a.question
	}

	doc := a.Doc()
	href, _ := doc.Find("h2.zm-item-title>a").Attr("href")
	link := makeZhihuLink(href)
	title := strip(doc.Find("h2.zm-item-title").First().Text())
	return NewQuestion(link, title)
}

// Author 返回该答案的作者
func (a *Answer) GetAuthor() *User {
	if a.author != nil {
		return a.author
	}

	doc := a.Doc()
	sel := doc.Find("div.zm-item-answer-author-info").First()
	if strip(sel.Text()) == "匿名用户" {
		return NewUser("", "匿名用户")
	}

	node := sel.Find("a.author-link")
	userId := strip(node.Text())
	urlPath, _ := node.Attr("href")
	userLink := makeZhihuLink(urlPath)
	return NewUser(userLink, userId)
}

// GetUpvote 返回赞同数
func (a *Answer) GetUpvote() int {
	if got, ok := a.fields["upvote"]; ok {
		return got.(int)
	}

	doc := a.Doc()
	text := strip(doc.Find("span.count").First().Text())
	upvote := 0
	if strings.HasSuffix(text, "K") {
		num, _ := strconv.Atoi(text[0 : len(text)-1])
		upvote = num * 1000
	} else if strings.HasPrefix(text, "W") {
		num, _ := strconv.Atoi(text[0 : len(text)-1])
		upvote = num * 10000
	} else {
		upvote, _ = strconv.Atoi(text)
	}
	a.fields["upvote"] = upvote
	return upvote
}

// TODO ToTxt 把回答导出到 txt 文件
func (a *Answer) ToTxt() error {
	return nil
}

// TODO ToMarkdown 把回答导出到 markdown 文件
func (a *Answer) ToMarkdown() error {
	return nil
}

// TODO GetContent 返回回答的内容
func (a *Answer) GetContent() string {
	return ""
}

// TODO GetVoters 返回点赞的用户
func (a *Answer) GetVoters() []*User {
	return nil
}

// TODO GetVisitTimes 返回所属问题被浏览次数
func (a *Answer) GetVisitTimes() int {
	return 0
}

func (a *Answer) String() string {
	return fmt.Sprintf("Answer: %s - %s", a.GetAuthor().String(), a.Link)
}
