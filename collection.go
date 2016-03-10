package zhihu

import (
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

// Collection 是一个知乎的收藏夹页面
type Collection struct {
	// Link 是该页面的链接
	Link string

	// doc 是一个 HTML document
	doc *goquery.Document

	// creator 是该收藏夹的创建者
	creator *User

	// name 是该收藏夹的名称
	name string

	// fields 是字段缓存，避免重复解析页面
	fields map[string]interface{}
}

func NewCollection(link string, name string) *Collection {
	if !validCollectionURL(link) {
		panic("收藏夹链接不正确：" + link)
	}

	return &Collection{
		Link:   link,
		name:   name,
		fields: make(map[string]interface{}),
	}
}

// Doc 用于获取当前问题页面的 HTML document，惰性求值
func (c *Collection) Doc() *goquery.Document {
	if c.doc != nil {
		return c.doc
	}

	var err error
	c.doc, err = newDocumentFromUrl(c.Link)
	if err != nil {
		return nil
	}

	return c.doc
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
	if got, ok := c.fields["followers-num"]; ok {
		return got.(int)
	}

	doc := c.Doc()

	// <a href="/collection/19653044/followers" data-za-c="collection" ,="" data-za-a="visit_collection_followers" data-za-l="collection_followers_count">
	//   7516
	// </a>
	text := strip(doc.Find(`a[data-za-a="visit_collection_followers"]`).Text())
	num, _ := strconv.Atoi(text)
	c.fields["followers-num"] = num
	return num
}

// TODO GetAllAnswers 返回收藏夹里所有的回答
func (c *Collection) GetAllAnswers() []*Answer {
	return nil
}

// TODO GetTopXAnswers 返回收藏夹里前 x 个回答
func (c *Collection) GetTopXAnswers(x int) []*Answer {
	return nil
}
