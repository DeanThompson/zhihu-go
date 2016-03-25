package zhihu

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Answer 是一个知乎的答案
type Answer struct {
	*Page

	// question 是该答案对应的问题
	question *Question

	// author 是该答案的作者
	author *User
}

// NewAnswer 用于创建一个 Answer 对象，其中 link 是必传的，question, author 可以为 nil
func NewAnswer(link string, question *Question, author *User) *Answer {
	return &Answer{
		Page:     newZhihuPage(link),
		question: question,
		author:   author,
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

// ToMarkdown 把回答导出到 markdown 文件
func (a *Answer) ToMarkdown(filename string) error {
	if !strings.HasSuffix(filename, ".md") && !strings.HasSuffix(filename, ".markdown") {
		filename += ".md"
	}

	// TODO convert to markdown
	md := ""

	return saveString(filename, md)
}

// ToHtml 把网页源码导出到 html 文件
func (a *Answer) ToHtml(filename string) error {
	if !strings.HasSuffix(filename, ".html") {
		filename += ".html"
	}

	html, err := a.Doc().Html()
	if err != nil {
		return err
	}
	return saveString(filename, html)
}

// GetContent 返回回答的内容，HTML 格式
func (a *Answer) GetContent() string {
	if got, ok := a.getStringField("content"); ok {
		return got
	}

	sel := a.Doc().Find("div#zh-question-answer-wrap").Find("div.zm-editable-content")
	content, err := answerSelectionToHtml(sel)
	if err != nil {
		logger.Error("导出 HTML 失败：%s", err.Error())
		return ""
	}
	a.setField("content", content)
	return content
}

// GetVotersN 返回 n 个点赞的用户，如果 n < 0，返回所有点赞的用户
func (a *Answer) GetVotersN(n int) []*User {
	if n == 0 {
		return nil
	}

	querystring := fmt.Sprintf(`params={"answer_id":"%d"}`, a.GetID())
	url := makeZhihuLink("/node/AnswerFullVoteInfoV2" + "?" + querystring)
	doc, err := newDocumentFromURL(url)
	if err != nil {
		return nil
	}

	sel := doc.Find(".voters span")
	capacity := n
	if capacity < 0 || capacity > sel.Length() {
		capacity = sel.Length()
	}
	voters := make([]*User, 0, capacity)

	sel.EachWithBreak(func(index int, span *goquery.Selection) bool {
		userId := strings.Trim(strip(span.Text()), "、")
		var userLink string
		if !(userId == "匿名用户" || userId == "知乎用户") {
			path, _ := span.Find("a").Attr("href")
			userLink = makeZhihuLink(path)
		}
		voters = append(voters, NewUser(userLink, userId))
		if n > 0 && len(voters) == n {
			return false
		}
		return true
	})

	return voters
}

// GetVoters 返回点赞的用户
func (a *Answer) GetVoters() []*User {
	return a.GetVotersN(-1)
}

// GetCommentsNum 返回评论数量
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

// GetCollectedNum 返回被收藏次数
func (a *Answer) GetCollectedNum() int {
	if value, ok := a.getIntField("collected-num"); ok {
		return value
	}

	text := strip(a.Doc().Find(`a[data-za-l="sidebar_answer_collected_count"]`).Text())
	value, _ := strconv.Atoi(text)
	a.setField("collected-num", value)
	return value
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

// 把一个回答的主体部分导出成 HTML 代码，与原码相比，做了这些操作：
// 	1. 去掉无用的 noscript 标签
// 	2. 修复 img 的 src 值
// 	3. 移除无用的 icon
// 	4. 如果是自己的回答，移除末尾的 “修改” 链接
func answerSelectionToHtml(sel *goquery.Selection) (string, error) {
	sel.RemoveClass()

	sel.Find("noscript").Each(func(_ int, tag *goquery.Selection) {
		tag.Remove() // 把无用的 noscript 去掉
	})

	sel.Find("i.icon-external").Each(func(_ int, tag *goquery.Selection) {
		tag.Remove() // 把无用的 icon 去掉
	})

	sel.Find("a.zu-edit-button").Remove() // 把 “修改” 链接去掉

	// 修复 img 的 src
	sel.Find("img").Each(func(_ int, tag *goquery.Selection) {
		var src string
		if tag.HasClass("origin_image") {
			src, _ = tag.Attr("data-original")
		} else {
			src, _ = tag.Attr("data-actualsrc")
		}
		tag.SetAttr("src", src)
		if tag.Next().Size() == 0 {
			tag.AfterHtml("<br>")
		}
	})

	// 修复 a 标签的 href，因为知乎的外链都是这种形式：https://link.zhihu.com/?target=xxx
	sel.Find("a").Each(func(_ int, tag *goquery.Selection) {
		href, _ := tag.Attr("href")
		if strings.Contains(href, "target=") {
			link, err := url.Parse(href)
			if err != nil {
				return
			}
			target := link.Query().Get("target")
			tag.SetAttr("href", target)
		}
	})

	wrapper := `<html><head><meta charset="utf-8"></head><body></body></html>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(wrapper))
	doc.Find("body").AppendSelection(sel)

	return doc.Html()
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
