package post_deleter

import (
	//"fmt"
	"regexp"
	"strings"

	"github.com/purstal/go-tieba-base/tieba"
	"github.com/purstal/go-tieba-base/tieba/adv-search"

	postfinder "github.com/purstal/go-tieba-modules/post-finder"
)

func (d *PostDeleter) ThreadFilter(account *postbar.Account, thread *postfinder.ForumPageThread) postfinder.Control {

	//fmt.Println(thread.Thread.LastReplyTime.Unix(), thread.Thread.Tid, thread.Thread.LastReplyer.ID)

	if (thread.Thread.Author.Name == "MC吧饮水姬" ||
		InStringSet(d.BawuList.KeyWords(), thread.Thread.Author.Name)) &&
		strings.Contains(thread.Thread.Title, "官方水楼") {
		if thread.Thread.LastReplyer.Name == "iamunknown" {
			return postfinder.Continue //测试用
		}
		d.Records.WaterThread_Tids[thread.Thread.Tid] = struct{}{}
		return postfinder.Finish
	} else if _, exist := d.Tid_Whitelist.KeyWords()[thread.Thread.Tid]; exist {
		//d.Logger.Debug("白名单内贴子")
		return postfinder.Finish
	}

	if InStringSet(d.BawuList.KeyWords(), thread.Thread.LastReplyer.Name) || InStringSet(d.UserName_Whitelist.KeyWords(), thread.Thread.LastReplyer.Name) {
		return postfinder.Finish
	}
	if InStringSet(d.BawuList.KeyWords(), thread.Thread.Author.Name) {
		if strings.Contains(thread.Thread.Title, "服务器发布贴") {
			d.Records.ServerListThread_Tids[thread.Thread.Tid] = struct{}{}
		} else if match, _ := regexp.MatchString(`吧规.*?\([0-9]*?.*?\)`, thread.Thread.Title); match {
			d.Records.RulesThread_Tids[thread.Thread.Tid] = struct{}{}
		} else if match, _ := regexp.MatchString(`基本守则`, thread.Thread.Title); match {
			d.Records.RulesThread_Tids[thread.Thread.Tid] = struct{}{}
		}

	}
	return postfinder.Continue
}

func (d *PostDeleter) PostAssessor(account *postbar.Account, post *postfinder.ThreadPagePost) {
	//logs.Debug(MakePrefix(GetServerTimeFromExtra(post.Extra), post.Thread.Tid, post.Post.Pid, 0, post.Post.Author.ID),
	//	"新回复") //, post.Thread.Title, post.Post.Author, post.Post.ContentList)
	/*
		if _, exist := d.Records.RulesThread_Tids[post.Thread.Tid]; exist &&
			!InStringSet(d.BawuList.KeyWords(), post.Post.Author.Name) {

			d.DeletePost("主题页面", account, &DeletePostRequest{
				tid:      post.Thread.Tid,
				pid:      post.Post.Pid,
				spid:     0,
				uid:      post.Post.Author.ID,
				title:    post.Thread.Title,
				content:  post.Post.ContentList,
				author:   post.Post.Author.Name,
				postTime: post.Post.PostTime.Format("2006-01-02 15:04:05"),
				reason:   "非吧务回复吧规",
				remark:   fmt.Sprintf("楼层:%d, ", post.Post.Floor),
			}, false, "")
			return
		}*/

	//DebugLog("一般回复", post.Post.PGetContentList())
	for _, content := range post.Post.ContentList {
		if link, ok := content.(postbar.Link); ok {
			if link.Text == "[语音]来自新版客户端语音功能" {
				d.Logger.Debug("有语音")
			}
		}
	}
	if d.CommonAssess("主题页面", account, post.Post, post.Thread) == postfinder.Finish {
		return
	}
}

func (d *PostDeleter) CommentAssessor(account *postbar.Account, comment *postfinder.FloorPageComment) {
	//logs.Debug(MakePrefix(GetServerTimeFromExtra(comment.Extra), comment.Thread.Tid, comment.Post.Pid, comment.Comment.Spid, comment.Comment.Author.ID),
	//	"新楼中楼回复") //, comment.Thread.Title, comment.Post.Author, comment.Comment.Author, comment.Comment.ContentList)
	if d.CommonAssess("楼层页面", account, comment.Comment, comment.Thread) == postfinder.Finish {
		return
	}
	//DebugLog("楼层回复", comment.Comment.PGetContentList())

}

func (d *PostDeleter) AdvSearchAssessor(account *postbar.Account, result *advsearch.AdvSearchResult) postfinder.Control {

	if _, exist := d.Records.WaterThread_Tids[result.Tid]; exist {
		if result.Author.Name == "iamunknown" {
			return postfinder.Continue //测试用.这样我在水楼的回复也能被找到并被处理...
		}
		return postfinder.Finish //防止水楼回复被删
	} /* else if _, exist := d.Records.RulesThread_Tids[result.Tid]; exist &&
		!InStringSet(d.BawuList.KeyWords(), result.Author.Name) {

		d.DeletePost("高级搜索", account, &DeletePostRequest{
			tid:      result.Tid,
			pid:      result.Pid,
			spid:     0,
			uid:      0,
			title:    result.Title,
			content:  result.Content,
			author:   result.Author.Name,
			postTime: result.PostTime.Format("2006-01-02 15:04"),
			reason:   "非吧务回复吧规",
			remark:   "",
		}, false, "")
		return postfinder.Finish
	}*/

	//DebugLog("高级搜索", result.PGetContentList())
	if len(result.Content) <= 120 {
		match, _ := regexp.MatchString(`回复.*?:`, result.Content)
		if match {
			return postfinder.Finish //疑似楼中楼而且内容完整的回复就不看了吧...
		}
	}

	if d.CommonAssess("高级搜索", account, result, advsearch.AdvSearchThread{Tid: result.Tid, Title: result.Title}) == postfinder.Finish {
		return postfinder.Finish
	}

	return postfinder.Continue
}
