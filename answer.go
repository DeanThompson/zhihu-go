package zhihu

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Answer 是一个知乎的答案
type Answer struct {
	*ZhihuPage

	// question 是该答案对应的问题
	question *Question

	// author 是该答案的作者
	author *User
}

// NewAnswer 用于创建一个 Answer 对象，其中 link 是必传的，question, author 可以为 nil
func NewAnswer(link string, question *Question, author *User) *Answer {
	return &Answer{
		ZhihuPage: newZhihuPage(link),
		question:  question,
		author:    author,
	}
}

// GetID 返回该答案的数字 ID
func (a *Answer) GetID() int {
	if got, ok := a.getIntField("data-aid"); ok {
		return got
	}

	doc := a.Doc()
	text, _ := doc.Find("div.zm-item-answer.zm-item-expanded").Attr("data-aid")
	aid, _ := strconv.Atoi(text)
	a.setField("data-aid", aid)
	return aid
}

// GetQuestion 返回该回答所属的问题，如果 NewAnswer 时 question 不为 nil，则直接返回该值；
// 否则会抓取页面并分析得到问题的链接和标题，再新建一个 Question 对象
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
	return newUserFromAnswerAuthorTag(sel)
}

// GetUpvote 返回赞同数
func (a *Answer) GetUpvote() int {
	if got, ok := a.getIntField("upvote"); ok {
		return got
	}

	doc := a.Doc()
	text := strip(doc.Find("span.count").First().Text())
	upvote := upvoteTextToNum(text)
	a.setField("upvote", upvote)
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

// GetContent 返回回答的内容
func (a *Answer) GetContent() string {
	if got, ok := a.getStringField("content"); ok {
		return got
	}

	doc := a.Doc()
	destDoc := goquery.CloneDocument(doc) // 从原文档 clone 一份，用于输出
	content := restructAnswerContent(destDoc, doc.Selection)
	a.setField("content", content)
	return content
}

// GetVoters 返回点赞的用户
func (a *Answer) GetVoters() []*User {
	querystring := fmt.Sprintf(`params={"answer_id":"%d"}`, a.GetID())
	url := makeZhihuLink("/node/AnswerFullVoteInfoV2" + "?" + querystring)
	doc, err := newDocumentFromUrl(url)
	if err != nil {
		return nil
	}

	sel := doc.Find(".voters span")
	voters := make([]*User, 0, sel.Length())
	sel.Each(func(index int, span *goquery.Selection) {
		userId := strings.Trim(strip(span.Text()), "、")
		var userLink string
		if !(userId == "匿名用户" || userId == "知乎用户") {
			path, _ := span.Find("a").Attr("href")
			userLink = makeZhihuLink(path)
		}
		voters = append(voters, NewUser(userLink, userId))
	})

	return voters
}

// GetVisitTimes 返回所属问题被浏览次数
func (a *Answer) GetVisitTimes() int {
	if got, ok := a.getIntField("visit-times"); ok {
		return got
	}

	doc := a.Doc()
	text := strip(doc.Find("div.zm-side-section.zh-answer-status p strong").Text())
	visitTimes, _ := strconv.Atoi(text)
	a.setField("visit-times", visitTimes)
	return visitTimes
}

func (a *Answer) GetCommentsNum() int {
	if value, ok := a.getIntField("comment-num"); ok {
		return value
	}

	doc := a.Doc()
	text := strip(doc.Find("a.meta-item.toggle-comment").Text())
	rv := reMatchInt(text)
	a.setField("comment-num", rv)
	return rv
}

func (a *Answer) String() string {
	return fmt.Sprintf("<Answer: %s - %s>", a.GetAuthor().String(), a.Link)
}

func (a *Answer) setContent(value string) {
	a.setField("content", value)
}

func (a *Answer) setUpvote(value int) {
	a.setField("upvote", value)
}

func upvoteTextToNum(text string) int {
	rv := 0
	if strings.HasSuffix(text, "K") {
		num, _ := strconv.Atoi(text[0 : len(text)-1])
		rv = num * 1000
	} else if strings.HasPrefix(text, "W") {
		num, _ := strconv.Atoi(text[0 : len(text)-1])
		rv = num * 10000
	} else {
		rv, _ = strconv.Atoi(text)
	}
	return rv
}

// TODO 可以重构一下
func restructAnswerContent(destDoc *goquery.Document, srcDoc *goquery.Selection) string {
	// 用于输出的 HTML，只保留页面的 header 部分，把 body 清空
	destDoc.Find("body").Children().Each(func(_ int, sel *goquery.Selection) {
		sel.Remove()
	})

	// 获取答案部分的 HTML，并进行清洗和修正
	answerSel := srcDoc.Find("div.zm-editable-content.clearfix")
	answerSel.Find("noscript").Each(func(_ int, sel *goquery.Selection) {
		sel.Remove() // 把无用的 noscript 部分去掉
	})

	answerSel.Find("img").Each(func(_ int, sel *goquery.Selection) {
		src, _ := sel.Attr("data-actualsrc")
		sel.SetAttr("src", src) // 把图片的 src 改为 data-actualsrc 的值
	})

	// body 只保留答案内容部分
	destDoc.Find("body").AppendSelection(answerSel)
	content, _ := destDoc.Html()
	return content
}

func newUserFromAnswerAuthorTag(sel *goquery.Selection) *User {
	if strip(sel.Text()) == "匿名用户" {
		return ANONYMOUS
	}

	node := sel.Find("a.author-link")
	userId := strip(node.Text())
	urlPath, _ := node.Attr("href")
	userLink := makeZhihuLink(urlPath)
	return NewUser(userLink, userId)
}
