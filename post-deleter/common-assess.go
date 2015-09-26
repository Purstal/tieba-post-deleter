package post_deleter

import (
	"fmt"
	"math"
	"regexp"

	"github.com/purstal/go-tieba-base/tieba"
	postfinder "github.com/purstal/go-tieba-modules/post-finder"
)

func (d *PostDeleter) CommonAssess(from string, account *postbar.Account, post postbar.IPost, thread postbar.IThread) postfinder.Control {

	_, uid := post.PGetAuthor().AGetID()
	pid := post.PGetPid()
	tid := thread.TGetTid()

	if _, exist := d.Records.WaterThread_Tids[tid]; exist {
		d.Logger.Debug(MakePrefix(nil, tid, pid, 0, uid), "水楼的贴子应该来不到这里,但是不知道为什么来了.")
		return postfinder.Finish //防止水楼回复被删
	} else if _, exist := d.Tid_Whitelist.KeyWords()[tid]; exist {
		d.Logger.Debug(MakePrefix(nil, tid, pid, 0, uid), "白名单内的贴子应该来不到这里,但是不知道为什么来了.")
		return postfinder.Finish
	} else if InStringSet(d.BawuList.KeyWords(), post.PGetAuthor().AGetName()) ||
		InStringSet(d.UserName_Whitelist.KeyWords(), post.PGetAuthor().AGetName()) {
		d.Logger.Debug(MakePrefix(nil, tid, pid, 0, uid), "白名单内的用户/吧务应该来不到这里,但是不知道为什么来了.")
		return postfinder.Finish
	}

	text := ExtractText(post.PGetContentList())

	var deleteReason, banReason []string

	if matchedExp := MatchAny(text, d.Content_RxKw.KeyWords()); matchedExp != nil {
		if matchedExp.BanFlag {
			banReason = append(banReason, fmt.Sprint("内容匹配关键词:", matchedExp))
		}
		deleteReason = append(deleteReason, fmt.Sprint("内容匹配关键词:", matchedExp))
	} else if math.Mod(float64(len(text)), 15.0) == 0 {
		if match, _ := regexp.MatchString("[1十拾⑩①][5五伍⑤]字", text); match {
			deleteReason = append(deleteReason, fmt.Sprint("标准十五字"))
		}
	}
	if matchedExp := MatchAny(post.PGetAuthor().AGetName(), d.UserName_RxKw.KeyWords()); matchedExp != nil {
		if matchedExp.BanFlag {
			banReason = append(banReason, fmt.Sprint("用户名匹配关键词:", matchedExp))
		}
		deleteReason = append(deleteReason, fmt.Sprint("用户名匹配关键词:", matchedExp))
	}

	var deleteReason_f string

	if len(deleteReason) != 0 {
		if len(deleteReason) == 1 {
			deleteReason_f = deleteReason[0]
		} else {
			deleteReason_f = fmt.Sprint(deleteReason)
		}
	}

	//TODO: 如果uid为0且需要封禁,重新获取uid

	ac, postTime := post.PGetPostTime()
	var postTime_str string
	if ac {
		postTime_str = postTime.Format("2006-01-02 15:04:05")
	} else {
		postTime_str = postTime.Format("2006-01-02 15:04")
	}

	if len(deleteReason) != 0 {
		d.DeletePost(from, account, &DeletePostRequest{
			tid:      tid,
			pid:      pid,
			spid:     0,
			uid:      uid,
			title:    thread.TGetTitle(),
			content:  post.PGetContentList(),
			author:   post.PGetAuthor().AGetName(),
			postTime: postTime_str,
			reason:   deleteReason_f,
			remark:   "",
		}, len(banReason) != 0, "")
	}

	return postfinder.Continue
}
