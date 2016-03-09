package zhihu

import "testing"

func init_session() {
	Init("./examples/config.json")
}

func Test_GetTitle(t *testing.T) {
	init_session()

	question := NewQuestion("https://www.zhihu.com/question/41171543", "")
	got := question.GetTitle()
	want := "如何评价第一局比赛 AlphaGo 战胜李世石？"
	logger.Info("got title: %s", got)
	logger.Info("expected title: %s", want)
	if got != want {
		t.Error("GetTitle() returns error result")
	}
}

func Test_GetDetail(t *testing.T) {
	init_session()

	question := NewQuestion("https://www.zhihu.com/question/41171543", "")
	got := question.GetDetail()
	want := "本题已收录至知乎圆桌 » 对弈人工智能，更多关于李世石对战人工智能的解读欢迎关注讨论。"
	logger.Info("got detail: %s", got)
	logger.Info("expected detail: %s", want)
	if got != want {
		t.Error("GetDetail() returns error result")
	}
}
