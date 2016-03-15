package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/DeanThompson/zhihu-go"
)

var (
	logger = zhihu.Logger{true}
)

func main() {
	zhihu.Init("./config.json")

	// 黄继新
	user := zhihu.NewUser("https://www.zhihu.com/people/jixin", "")
	showUser(user)

	logger.Success("========== split ==========")

	// 龙有九个儿子，是跟谁生的？为什么「龙生九子，各不成龙」？
	questionUrl := "https://www.zhihu.com/question/23759686"
	question := zhihu.NewQuestion(questionUrl, "")
	showQuestion(question)

	logger.Success("========== split ==========")

	// 龙有九个儿子，是跟谁生的？为什么「龙生九子，各不成龙」？豆子 的答案
	answer := zhihu.NewAnswer("https://www.zhihu.com/question/23759686/answer/41997389", nil, nil)
	showAnswer(answer)

	logger.Success("========== split ==========")

	// 程序员为了期权加入创业公司，值得吗？ 匿名用户的答案
	answer2 := zhihu.NewAnswer("https://www.zhihu.com/question/28023819/answer/49723406", nil, nil)
	showAnswer(answer2)

	logger.Success("========== split ==========")

	collection := zhihu.NewCollection("https://www.zhihu.com/collection/19653044", "")
	showCollection(collection)
}

func showQuestion(question *zhihu.Question) {
	logger.Info("Question fields:")
	logger.Info("	url: %s", question.Link)
	logger.Info("	title: %s", question.GetTitle())
	logger.Info("	detail: %s", question.GetDetail())
	logger.Info("	answers num: %d", question.GetAnswersNum())
	logger.Info("	followers num: %d", question.GetFollowersNum())
	logger.Info("	topics: %s", strings.Join(question.GetTopics(), ", "))

	allAnswers := question.GetAllAnswers()
	for i, answer := range allAnswers {
		logger.Info("	answer-%d: %s", i+1, answer.String())
		// filename := fmt.Sprintf("/tmp/%s-%s的回答.html", question.GetTitle(), answer.GetAuthor().GetUserID())
		// dumpAnswerHTML(filename, answer)
		if i >= 10 {
			logger.Info("	%d answers not shown.", len(allAnswers)-i-1)
			break
		}
	}

	topXAnswers := question.GetTopXAnswers(25)
	for i, answer := range topXAnswers {
		logger.Info("	top-%d answer: %s", i+1, answer.String())
	}

	logger.Info("	top-1 answer: %s", question.GetTopAnswer().String())
	logger.Info("	visit times: %d", question.GetVisitTimes())
}

func showAnswer(answer *zhihu.Answer) {
	return
	logger.Info("Answer fields:")
	logger.Info("	url: %s", answer.Link)

	question := answer.GetQuestion()
	logger.Info("	question url: %s", question.Link)
	logger.Info("	question title: %s", question.GetTitle())

	logger.Info("	author: %s", answer.GetAuthor().String())
	logger.Info("	upvote num: %d", answer.GetUpvote())
	logger.Info("	visit times: %d", answer.GetVisitTimes())
	logger.Info("	data ID: %d", answer.GetID())

	// dump content
	filename := fmt.Sprintf("/tmp/answer_%d.html", answer.GetID())
	dumpAnswerHTML(filename, answer)

	for i, voter := range answer.GetVoters() {
		logger.Info("	voter-%d: %s", i+1, voter.String())
		if i >= 10 {
			remain := len(answer.GetVoters()) - i - 1
			logger.Info("	%d votes not shown.", remain)
			break
		}
	}
}

func showCollection(collection *zhihu.Collection) {
	logger.Info("Collection fields:")
	logger.Info("	url: %s", collection.Link)
	logger.Info("	name: %s", collection.GetName())
	logger.Info("	creator: %s", collection.GetCreator().String())
	logger.Info("	followers num: %d", collection.GetFollowersNum())
}

func showUser(user *zhihu.User) {
	logger.Info("User fields:")
	logger.Info("	is anonymous: %v", user.IsAnonymous())
	logger.Info("	userId: %s", user.GetUserID())
	logger.Info("	dataId: %s", user.GetDataID())
	logger.Info("	bio: %s", user.GetBio())
	logger.Info("	location: %s", user.GetLocation())
	logger.Info("	business: %s", user.GetBusiness())
	logger.Info("	gender: %s", user.GetGender())
	logger.Info("	followers num: %d", user.GetFollowersNum())
	logger.Info("	followees num: %d", user.GetFolloweesNum())
	logger.Info("	agree num: %d", user.GetAgreeNum())
	logger.Info("	thanks num: %d", user.GetThanksNum())
	logger.Info("	asks num: %d", user.GetAsksNum())
	logger.Info("	answers num: %d", user.GetAnswersNum())
	logger.Info("	posts num: %d", user.GetPostsNum())
	logger.Info("	collections num: %d", user.GetCollectionsNum())
	logger.Info("	logs num: %d", user.GetLogsNum())

	for i, follower := range user.GetFollowers() {
		logger.Info("	follower-%d: %s", i, follower.String())
	}

	for i, followee := range user.GetFollowers() {
		logger.Info("	followee-%d: %s", i, followee.String())
	}

	for i, ask := range user.GetAsks() {
		logger.Info("	ask-%d: %s", i, ask.String())
	}

	for i, answer := range user.GetAnswers() {
		logger.Info("	answer-%d: %s", i, answer.String())
	}

	for i, collection := range user.GetCollections() {
		logger.Info("	collection-%d: %s", i, collection.String())
	}

	for i, like := range user.GetLikes() {
		logger.Info("	like-%d: %s", i, like.String())
	}
}

func dumpAnswerHTML(filename string, answer *zhihu.Answer) error {
	fd, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		var _err error
		if os.IsNotExist(err) {
			fd, _err = os.Create(filename)
			if _err != nil {
				return _err
			}
		}
	}
	_, err = fd.WriteString(answer.GetContent())
	if err == nil {
		logger.Info("	content dumped to %s", filename)
	}
	return err
}
