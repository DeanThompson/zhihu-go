zhihu-go：知乎非官方 API 库 with Go
=================================

[![GoDoc](https://godoc.org/github.com/DeanThompson/zhihu-go?status.svg)](https://godoc.org/github.com/DeanThompson/zhihu-go)

这是一个非官方的 [知乎](https://www.zhihu.com/) API 库，用 Go 实现。

本项目基本上是把 [zhihu-python](https://github.com/egrcc/zhihu-python) 和 [zhihu-py3](https://github.com/7sDream/zhihu-py3) 从 Python 移植到了 Go. 相比之下，比 zhihu-python 的 API 更丰富，比 zhihu-py3 少了活动相关的 API.

**注意：知乎的 API、前端等都可能随时会更新，所以本项目的接口可能会有过时的情况。如果遇到此类问题，欢迎提交 issue 或 pull requests.**

## Table of Contents

* [Table of Contents](#table-of-contents)
* [Install](#install)
* [Documentation](#documentation)
* [Usage](#usage)
  * [Login：登录](#login)
  * [User：获取用户信息](#user)
  * [Question：获取问题信息](#question)
  * [Answer：获取答案信息](#answer)
  * [Collection：获取收藏夹信息](#collection)
  * [Topic：获取话题信息](#topic)
* [Known Issues](#known-issues)
* [TODO](#todo)
* [LICENSE](#license)

## Install

直接使用 `go get`:

```bash
go get github.com/DeanThompson/zhihu-go
```

依赖以下第三方库：

* [goquery](https://github.com/PuerkitoBio/goquery)： 用于解析 HTML，语法操作类似 jQuery
* [color](https://github.com/fatih/color)：用于输出带颜色的日志
* [persistent-cookiejar](https://github.com/juju/persistent-cookiejar)：用于维护一个持久化的 cookiejar，实现保持登录

## Documentation

请点击链接前往 GoDoc 查看：[zhihu-go](https://godoc.org/github.com/DeanThompson/zhihu-go)

## Usage

目前已经实现了用户（User），问题（Question），回答（Answer），收藏夹（Collection），话题（Topic）相关的 API，都是信息获取类的，暂无操作类的。

zhihu-go 包名为 `zhihu`，使用前需要先 import:

```go
import "github.com/DeanThompson/zhihu-go"
```

### Login

调用 API 之前需要先登录。在 zhihu-go 内部，使用一个全局的 session 来访问所有页面，并自动处理 cookies.

创建一个 JSON 格式的配置文件，提供一个账号和密码，格式如 [config-example.json](examples/config-example.json).

登录（初始化 session）：

```go
zhihu.Init("/path/to/config.json")
```

第一次登录会调用图像界面打开验证码文件，需要手动输入验证码到控制台。如果登录成功，后续的请求会沿用此次登录的 cookie, 不需要重复登录。

### User

`zhihu.User` 表示一个知乎用户，可以用于获取一个用户的各种数据。

创建一个 `User` 对象需要传入用户主页的 URL 及其知乎 ID（用户名），如：

```go
link := "https://www.zhihu.com/people/jixin"
userID := "黄继新"
user := zhihu.NewUser(link, userID)
```

获取用户的数据（代码见：[example.go](examples/example.go#L159)）：

```go
func showUser(user *zhihu.User) {
	logger.Info("User fields:")
	logger.Info("	is anonymous: %v", user.IsAnonymous())  // 是否匿名用户：false
	logger.Info("	userId: %s", user.GetUserID())          // 知乎ID：黄继新
	logger.Info("	dataId: %s", user.GetDataID())          // hash ID：b6f80220378c8b0b78175dd6a0b9c680
	logger.Info("	bio: %s", user.GetBio())                // BIO：和知乎在一起
	logger.Info("	location: %s", user.GetLocation())      // 位置：北京
	logger.Info("	business: %s", user.GetBusiness())      // 行业：互联网
	logger.Info("	gender: %s", user.GetGender())          // 性别：male
	logger.Info("	education: %s", user.GetEducation())    // 学校：北京第二外国语学院
	logger.Info("	followers num: %d", user.GetFollowersNum()) // 粉丝数：756632
	logger.Info("	followees num: %d", user.GetFolloweesNum()) // 关注的人数： 9249
	logger.Info("	followed columns num: %d", user.GetFollowedColumnsNum()) // 关注的专栏数：631
	logger.Info("	followed topics num: %d", user.GetFollowedTopicsNum())   // 关注的话题数：131
	logger.Info("	agree num: %d", user.GetAgreeNum())     // 获得的赞同数：68557
	logger.Info("	thanks num: %d", user.GetThanksNum())   // 获得的感谢数：17651
	logger.Info("	asks num: %d", user.GetAsksNum())       // 提问数：1336
	logger.Info("	answers num: %d", user.GetAnswersNum()) // 回答数：785
	logger.Info("	posts num: %d", user.GetPostsNum())     // 专栏文章数：92
	logger.Info("	collections num: %d", user.GetCollectionsNum()) // 收藏夹数量：44
	logger.Info("	logs num: %d", user.GetLogsNum())   // 公共编辑数：51596
	
	// <Topic: 知乎指南 - https://www.zhihu.com/topic/19550235>
	// <Topic: 苹果公司 (Apple Inc.) - https://www.zhihu.com/topic/19551762>
	// <Topic: 创新工场 - https://www.zhihu.com/topic/19624098>
	// <Topic: iPhone - https://www.zhihu.com/topic/19550292>
	// <Topic: 风险投资（VC） - https://www.zhihu.com/topic/19550422>
	for i, topic := range user.GetFollowedTopicsN(5) {
		logger.Info("	top followed topic-%d: %s", i+1, topic.String())
	}

	// <User: Zz XI - https://www.zhihu.com/people/zz-xi-18>
	// <User: xyn - https://www.zhihu.com/people/xyn-31>
	// <User: 江湖人称丸子头 - https://www.zhihu.com/people/jiang-hu-ren-cheng-wan-zi-tou>
	// <User: 小萍果Y - https://www.zhihu.com/people/xiao-ping-guo-y>
	// <User: 最爱麦丽素 - https://www.zhihu.com/people/Mylikes-82>
	for i, follower := range user.GetFollowersN(5) {
		logger.Info("	top follower-%d: %s", i+1, follower.String())
	}

	// <User: 最爱麦丽素 - https://www.zhihu.com/people/Mylikes-82>
	// <User: meidong - https://www.zhihu.com/people/zhalimuto>
	// <User: 青锐吴斌 - https://www.zhihu.com/people/wu-bin-817>
	// <User: Klaith - https://www.zhihu.com/people/Klaith>
	// <User: 张野 - https://www.zhihu.com/people/zhang-ye-91-9>
	for i, followee := range user.GetFolloweesN(5) {
		logger.Info("	top followee-%d: %s", i+1, followee.String())
	}

	// <Question: 偏好投票制（Preferential Voting）的优点和缺点是什么？最适用于哪类场合？ - https://www.zhihu.com/question/40939579>
	// <Question: 关于飞机上对使用手机的限制，为什么国内航班比国外航班严？ - https://www.zhihu.com/question/34302422>
	// <Question: 银联人民币卡可以在美国支持银联的 ATM 机上直接取美元吗？ - https://www.zhihu.com/question/33874729>
	// <Question: 小猫掉进了 5 米深的天井，如何能尽快救出？救助时应注意什么？ - https://www.zhihu.com/question/33307041>
	// <Question: 一件商品打一折（90% off）销售，这属于「超高折扣」还是「超低折扣」？ - https://www.zhihu.com/question/31332557>
	for i, ask := range user.GetAsksN(5) {
		logger.Info("	top ask-%d: %s", i+1, ask.String())
	}

	// <Answer: <User: 黄继新 - https://www.zhihu.com/people/jixin> - https://www.zhihu.com/question/40394171/answer/86692178>
	// <Answer: <User: 黄继新 - https://www.zhihu.com/people/jixin> - https://www.zhihu.com/question/19952708/answer/84561308>
	// <Answer: <User: 黄继新 - https://www.zhihu.com/people/jixin> - https://www.zhihu.com/question/35987345/answer/72981016>
	// <Answer: <User: 黄继新 - https://www.zhihu.com/people/jixin> - https://www.zhihu.com/question/24980451/answer/29789141>
	// <Answer: <User: 黄继新 - https://www.zhihu.com/people/jixin> - https://www.zhihu.com/question/24816698/answer/29229733>
	for i, answer := range user.GetAnswersN(5) {
		logger.Info("	top answer-%d: %s", i+1, answer.String())
	}

	// <Collection: 单子 - https://www.zhihu.com/collection/36510307>
	// <Collection: 稍后回答 - https://www.zhihu.com/collection/19665350>
	// <Collection: 广告！ - https://www.zhihu.com/collection/19688005>
	// <Collection: 关于知乎的思考 - https://www.zhihu.com/collection/19573315>
	// <Collection: MD，说得太好了！ - https://www.zhihu.com/collection/19886553>
	for i, collection := range user.GetCollectionsN(5) {
		logger.Info("	top collection-%d: %s", i+1, collection.String())
	}

	for i, like := range user.GetLikes() {
		logger.Info("	like-%d: %s", i+1, like.String())
	}
}
```

### Question

`zhihu.Question` 表示一个知乎问题，用于获取问题相关的数据。初始化需要提供 url 和标题（可为空）:

```go
link := "https://www.zhihu.com/question/28966220"
title := "Python 编程，应该养成哪些好的习惯？"
question := zhihu.NewQuestion(link, title)
```

获取问题数据：（代码见：[example.go](examples/example.go#L51)）

```go
func showQuestion(question *zhihu.Question) {
	logger.Info("Question fields:")
	
	// 链接：https://www.zhihu.com/question/28966220
	logger.Info("	url: %s", question.Link)
	
	// 标题：Python 编程，应该养成哪些好的习惯？
	logger.Info("	title: %s", question.GetTitle())
	
	// 描述：我以为编程习惯很重要的，一开始就养成这些习惯，不仅可以提高编程速度，还可以减少 bug 出现的概率。希望各位分享好的编程习惯。
	logger.Info("	detail: %s", question.GetDetail())
	
	
	logger.Info("	answers num: %d", question.GetAnswersNum()) // 回答数：15
	logger.Info("	followers num: %d", question.GetFollowersNum()) // 关注者数量：1473

	// <Topic: 程序员 - https://www.zhihu.com/topic/19552330>
	// <Topic: Python - https://www.zhihu.com/topic/19552832>
	// <Topic: 编程 - https://www.zhihu.com/topic/19554298>
	// <Topic: Python 入门 - https://www.zhihu.com/topic/19661050>
	for i, topic := range question.GetTopics() {
		logger.Info("	topic-%d: %s", i+1, topic.String())
	}

	// <User: 铁头爸爸 - https://www.zhihu.com/people/li-liang-68-9>
	// <User: 阳阳 - https://www.zhihu.com/people/yang-yang-3-29-52>
	// <User: 田小芳 - https://www.zhihu.com/people/tian-xiao-fang-55>
	// <User: 濕濕 - https://www.zhihu.com/people/shi-shi-29-7-18>
	// <User: 陈翔宇 - https://www.zhihu.com/people/chen-xiang-yu-91-74>
	for i, follower := range question.GetFollowersN(5) {
		logger.Info("	top follower-%d: %s", i+1, follower.String())
	}

	for i, follower := range question.GetFollowers() {  // 关注者列表
		logger.Info("	follower-%d: %s", i+1, follower.String())
		if i >= 10 {
			logger.Info("	%d followers not shown.", question.GetFollowersNum()-i-1)
			break
		}
	}

	allAnswers := question.GetAllAnswers()  // 所有回答
	for i, answer := range allAnswers {
		logger.Info("	answer-%d: %s", i+1, answer.String())
		filename := fmt.Sprintf("/tmp/%s-%s的回答.html", question.GetTitle(), answer.GetAuthor().GetUserID())
		dumpAnswerHTML(filename, answer)
		if i >= 10 {
			logger.Info("	%d answers not shown.", len(allAnswers)-i-1)
			break
		}
	}

	topXAnswers := question.GetTopXAnswers(25)  // 前 25 个回答
	for i, answer := range topXAnswers {
		logger.Info("	top-%d answer: %s", i+1, answer.String())
	}

	// 排名第一的回答
	// <Answer: <User: 陈村 - https://www.zhihu.com/people/xjiangxjxjxjx> - https://www.zhihu.com/question/28966220/answer/43346747>
	logger.Info("	top-1 answer: %s", question.GetTopAnswer().String())
	
	logger.Info("	visit times: %d", question.GetVisitTimes()) // 查看次数：32942
}
```

### Answer

`zhihu.Answer` 表示一个知乎答案，初始化时需要指定页面链接，也支持指定对应的问题（`*Question`，可以为 `nil`）和作者（`*User`，可以为 `nil`）：

```go
// 龙有九个儿子，是跟谁生的？为什么「龙生九子，各不成龙」？豆子 的答案
answer := zhihu.NewAnswer("https://www.zhihu.com/question/23759686/answer/41997389", nil, nil)
```

获取回答数据：（代码见：[example.go](examples/example.go#L95)）

```go
func showAnswer(answer *zhihu.Answer) {
	logger.Info("Answer fields:")
	
	// 链接：https://www.zhihu.com/question/23759686/answer/41997389
	logger.Info("	url: %s", answer.Link)

	// 所属问题
	// 链接：https://www.zhihu.com/question/23759686
	// 标题：龙有九个儿子，是跟谁生的？为什么「龙生九子，各不成龙」？
	question := answer.GetQuestion()
	logger.Info("	question url: %s", question.Link)
	logger.Info("	question title: %s", question.GetTitle())

	// 作者：<User: 豆子 - https://www.zhihu.com/people/douzishushu>
	logger.Info("	author: %s", answer.GetAuthor().String())
	
	logger.Info("	upvote num: %d", answer.GetUpvote())    // 赞同数：26486
	logger.Info("	comments num: %d", answer.GetCommentsNum()) // 评论数：20
	logger.Info("	collected num: %d", answer.GetCollectedNum())	// 被收藏次数：22929
	logger.Info("	data ID: %d", answer.GetID())   // 数字 ID：12191779

	// 点赞的用户
	voters := answer.GetVoters()
	for i, voter := range voters {
		logger.Info("	voter-%d: %s", i+1, voter.String())
		if i >= 10 {
			remain := len(voters) - i - 1
			logger.Info("	%d votes not shown.", remain)
			break
		}
	}
}
```

### Collection

`zhihu.Collection` 表示一个收藏夹，初始化时必须指定页面 url，支持指定名称（`string` 可以为 `""`）和创建者（`creator *User`，可以为 `nil`）：

```go
// 黄继新 A4U
collection := zhihu.NewCollection("https://www.zhihu.com/collection/19677733", "", nil)
```

获取收藏夹数据：（代码见：[example.go](examples/example.go#L124)）

```go
func showCollection(collection *zhihu.Collection) {
	logger.Info("Collection fields:")
	
	// 链接：https://www.zhihu.com/collection/19677733
	logger.Info("	url: %s", collection.Link)
	
	// 名称：A4U
	logger.Info("	name: %s", collection.GetName())
	
	// 作者：<User: 黄继新 - https://www.zhihu.com/people/jixin>
	logger.Info("	creator: %s", collection.GetCreator().String())
	logger.Info("	followers num: %d", collection.GetFollowersNum())   // 关注者数量：29

	// 获取 5 个关注者
	for i, follower := range collection.GetFollowersN(5) {
		logger.Info("	top follower-%d: %s", i+1, follower.String())
	}
	
	// 获取 5 个问题
	for i, question := range collection.GetQuestionsN(5) {
		logger.Info("	top question-%d: %s", i+1, question.String())
	}

	// 获取 5 个回答
	for i, answer := range collection.GetAnswersN(5) {
		logger.Info("	top answer-%d: %s", i+1, answer.String())
	}
}
```

### Topic

`zhihu.Collection` 表示一个话题，初始化时必须指定页面 url，支持指定名称（`string` 可以为 `""`）：

```go
// Python
topic := zhihu.NewTopic("https://www.zhihu.com/topic/19552832", "")
```

获取收藏夹数据：（代码见：[example.go](examples/example.go#L237)）

```go
func showTopic(topic *zhihu.Topic) {
	logger.Info("Topic fields:")
	
	// 链接：https://www.zhihu.com/topic/19552832
	logger.Info("	url: %s", topic.Link)
	
	// 名称：Python
	logger.Info("	name: %s", topic.GetName())
	
	// 描述：Python 是一种面向对象的解释型计算机程序设计语言，在设计中注重代码的可读性，同时也是一种功能强大的通用型语言。
	logger.Info("	description: %s", topic.GetDescription())
	
	// 关注者数量：82805
	logger.Info("	followers num: %d", topic.GetFollowersNum())

	// 最佳答主，一般为 5 个
	// <User: RednaxelaFX - https://www.zhihu.com/people/rednaxelafx>
	// <User: 松鼠奥利奥 - https://www.zhihu.com/people/tonyseek>
	// <User: 涛吴 - https://www.zhihu.com/people/Metaphox>
	// <User: 冯昱尧 - https://www.zhihu.com/people/feng-yu-yao>
	// <User: Coldwings - https://www.zhihu.com/people/coldwings>
	for i, author := range topic.GetTopAuthors() {
		logger.Info("	top-%d author: %s", i+1, author.String())
	}
}
```

## Known Issues

无，欢迎 [提交 issues](https://github.com/DeanThompson/zhihu-go/issues)

## TODO

按优先级降序排列：

* [X] 获取回答的收藏数
* [X] 获取收藏夹的答案数量
* [X] 获取用户的头像
* [X] 获取用户的微博地址
* [ ] 把答案导出到 markdown 文件
* [ ] 更多的登录方式，不需要依赖图形界面打开验证码文件
* [ ] 增加评论相关的 API
* [ ] 增加活动相关的 API
* [ ] 增加专栏相关的 API
* [ ] test（暂时没想好怎么做）

很可能不会做：

* [ ] 增加用户的操作，如点赞、关注等

欢迎 [提交 pull requests](https://github.com/DeanThompson/zhihu-go/pulls)

## LICENSE

[The MIT license](LICENSE).
