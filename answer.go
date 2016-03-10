package zhihu

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Answer 是一个知乎的答案
type Answer struct {
	// Link 是该答案的链接
	Link string

	// doc 是一个 HTML document
	doc *goquery.Document

	// question 是该答案对应的问题
	question *Question

	// author 是该答案的作者
	author *User

	// fields 是一些其他信息的缓存
	fields map[string]interface{}
}

// NewAnswer 用于创建一个 Answer 对象，其中 link 是必传的，question, author 可以为 nil
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

	var err error
	a.doc, err = newDocumentFromUrl(a.Link)
	if err != nil {
		return nil
	}

	return a.doc
}

// GetID 返回该答案的数字 ID
func (a *Answer) GetID() int {
	if got, ok := a.fields["data-aid"]; ok {
		return got.(int)
	}

	var (
		doc = a.Doc()
		aid = 0
	)
	text, exists := doc.Find("div.zm-item-answer.zm-item-expanded").Attr("data-aid")
	if exists {
		aid, _ = strconv.Atoi(text)
	}
	a.fields["data-aid"] = aid
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

// GetContent 返回回答的内容
func (a *Answer) GetContent() string {
	if got, ok := a.fields["content"]; ok {
		return got.(string)
	}

	doc := a.Doc()

	// 从原文档 clone 一份
	newDoc := goquery.CloneDocument(doc)

	// 把 body 清空
	newDoc.Find("body").Children().Each(func(_ int, sel *goquery.Selection) {
		sel.Remove()
	})

	// 获取答案部分的 HTML，并进行清洗和修正
	answerSel := doc.Find("div.zm-editable-content.clearfix")
	answerSel.Find("noscript").Each(func(_ int, sel *goquery.Selection) {
		sel.Remove() // 把无用的 noscript 部分去掉
	})

	answerSel.Find("img").Each(func(_ int, sel *goquery.Selection) {
		src, _ := sel.Attr("data-actualsrc")
		sel.SetAttr("src", src) // 把图片的 src 改为 data-actualsrc 的值
	})

	// body 只保留答案内容部分
	newDoc.Find("body").AppendSelection(answerSel)
	content, _ := newDoc.Html()
	a.fields["content"] = content
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
	if got, ok := a.fields["visit-times"]; ok {
		return got.(int)
	}

	doc := a.Doc()
	text := strip(doc.Find("div.zm-side-section.zh-answer-status p strong").Text())
	visitTimes, _ := strconv.Atoi(text)
	a.fields["visit-times"] = visitTimes
	return visitTimes
}

func (a *Answer) String() string {
	return fmt.Sprintf("<Answer: %s - %s>", a.GetAuthor().String(), a.Link)
}
