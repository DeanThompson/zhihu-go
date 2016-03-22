package zhihu

import (
	"fmt"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

type Topic struct {
	*ZhihuPage

	// name 是改话题的名称
	name string
}

func NewTopic(link string, name string) *Topic {
	if !validTopicURL(link) {
		panic("非法的 Topic 链接：%s" + link)
	}
	return &Topic{
		ZhihuPage: newZhihuPage(link),
		name:      name,
	}
}

// GetName 返回话题名称
func (t *Topic) GetName() string {
	if t.name != "" {
		return t.name
	}

	// <h1 class="zm-editable-content" data-disabled="1">Python</h1>
	t.name = strip(t.Doc().Find("h1.zm-editable-content").Text())
	return t.name
}

// GetDescription 返回话题的描述
func (t *Topic) GetDescription() string {
	if got, ok := t.getStringField("description"); ok {
		return got
	}

	// <div class="zm-editable-content" data-editable-maxlength="130">
	//   Python 是一种面向对象的解释型计算机程序设计语言，在设计中注重代码的可读性，同时也是一种功能强大的通用型语言。
	//   <a href="javascript:;" class="zu-edit-button" name="edit">
	//     <i class="zu-edit-button-icon"></i>修改
	//   </a>
	// </div>
	description := strip(t.Doc().Find("div.zm-editable-content").Text())
	t.setField("description", description)
	return description
}

// GetFollowersNum 返回关注者数量
func (t *Topic) GetFollowersNum() int {
	if got, ok := t.getIntField("followers-num"); ok {
		return got
	}

	// <div class="zm-topic-side-followers-info">
	// 	 <a href="/topic/19552832/followers">
	//     <strong>82155</strong>
	//   </a> 人关注了该话题
	// </div>
	text := strip(t.Doc().Find("div.zm-topic-side-followers-info strong").Text())
	num, _ := strconv.Atoi(text)
	t.setField("followers-num", num)
	return num
}

// GetTopAuthors 返回最佳回答者，一般来说是 5 个
func (t *Topic) GetTopAuthors() []*User {
	authors := make([]*User, 0, 5)
	div := t.Doc().Find("div#zh-topic-top-answerer")
	div.Find("div.zm-topic-side-person-item-content").Each(func(index int, sel *goquery.Selection) {
		uHref, _ := sel.Find("a").Attr("href")
		uId := strip(sel.Find("a").Text())

		thisAuthor := NewUser(makeZhihuLink(uHref), uId)

		bio, _ := sel.Find("div.zm-topic-side-bio").Attr("title")
		thisAuthor.setBio(bio)

		authors = append(authors, thisAuthor)
	})
	return authors
}

func (t *Topic) String() string {
	return fmt.Sprintf("<Topic: %s - %s>", t.GetName(), t.Link)
}
