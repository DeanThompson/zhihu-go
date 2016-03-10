package main

import (
	"strings"

	"github.com/DeanThompson/zhihu-go"
)

var (
	logger = zhihu.Logger{true}
)

func main() {
	zhihu.Init("./config.json")

	// 如何评价第一局比赛 AlphaGo 战胜李世石？
	questionUrl := "https://www.zhihu.com/question/41171543"
	question := zhihu.NewQuestion(questionUrl, "")
	showQuestion(question)

	logger.Success("========== split ==========")

	// 龙有九个儿子，是跟谁生的？为什么「龙生九子，各不成龙」？豆子 的答案
	answerUrl := "https://www.zhihu.com/question/23759686/answer/41997389"
	answer := zhihu.NewAnswer(answerUrl, nil, nil)
	showAnswer(answer)

	logger.Success("========== split ==========")

	// 程序员为了期权加入创业公司，值得吗？ 匿名用户的答案
	answer2 := zhihu.NewAnswer("https://www.zhihu.com/question/28023819/answer/49723406", nil, nil)
	showAnswer(answer2)
}

func showQuestion(question *zhihu.Question) {
	logger.Info("Question fields:")
	logger.Info("	url: %s", question.Link)
	logger.Info("	title: %s", question.GetTitle())
	logger.Info("	detail: %s", question.GetDetail())
	logger.Info("	answers num: %d", question.GetAnswersNum())
	logger.Info("	followers num: %d", question.GetFollowersNum())
	logger.Info("	topics: %s", strings.Join(question.GetTopics(), ", "))

	//	for i, answer := range question.GetAllAnswers() {
	//		logger.Info("	answer-%d: %s", i, answer.String())
	//	}
	//
	//	logger.Info("	top-1 answer: %s", question.GetTopAnswer().String())
	logger.Info("	visit times: %d", question.GetVisitTimes())
}

func showAnswer(answer *zhihu.Answer) {
	logger.Info("Answer fields:")
	logger.Info("	url: %s", answer.Link)

	question := answer.GetQuestion()
	//	showQuestion(answer.GetQuestion())
	logger.Info("	question url: %s", question.Link)
	logger.Info("	question title: %s", question.GetTitle())

	logger.Info("	author: %s", answer.GetAuthor().String())
	logger.Info("	upvote num: %d", answer.GetUpvote())
	logger.Info("	visit times: %d", answer.GetVisitTimes())
	logger.Info("	data ID: %d", answer.GetID())

	for i, voter := range answer.GetVoters() {
		logger.Info("	voter-%d: %s", i, voter.String())
	}
}
