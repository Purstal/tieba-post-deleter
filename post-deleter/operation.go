package post_deleter

import (
	"fmt"
	"time"

	"github.com/purstal/go-tieba-base/tieba"
	"github.com/purstal/go-tieba-base/tieba/apis"
)

type DeletePostRequest struct {
	tid, pid, spid, uid uint64

	title    string
	content  interface{}
	author   string
	postTime string

	reason string
	remark string
}

func (d *PostDeleter) DeletePost(from string, account *postbar.Account, r *DeletePostRequest, withBan bool, givedBanReason string) {

	if withBan {
		if givedBanReason == "" {
			givedBanReason = "null"
		}

		d.BanID(from, account.BDUSS, &BanIDRequest{
			tid:          r.tid,
			pid:          r.pid,
			spid:         r.spid,
			uid:          r.uid,
			day:          1,
			userName:     r.author,
			title:        r.title,
			content:      r.content,
			postTime:     r.postTime,
			loggedReason: r.reason,
			givedReason:  givedBanReason,
		})
	}

	if account.BDUSS == "" {
		d.OpLogger.Record(from, time.Now(), r, "忽略 未设置BDUSS")
		return
	}

	var op_pid uint64
	if r.spid != 0 {
		op_pid = r.spid
	} else {
		op_pid = r.pid
	}

	var status string
	now := time.Now()

	for i := 0; ; i++ {
		err, pberr := apis.DeletePost(account, op_pid)
		if err == nil && (pberr == nil || pberr.ErrorCode == 0) {
			status = "成功"
			break
		} else if i < 3 {
			d.Logger.Error("删贴失败,将最多尝试三次:", err, pberr, ".")
		} else {
			d.Logger.Error("删贴失败,放弃:", err, pberr, ".")
			if err != nil {
				status = fmt.Sprint("失败", err)
			} else {
				status = fmt.Sprint("失败", pberr)
			}
			break
		}
	}

	d.OpLogger.Record(from, now, r, status)

}

type DeleteThreadRequest struct {
	tid, uid, pid uint64
	title         string
	content       interface{}
	author        string
	reason        string
	postTime      string
	remark        string
}

func (d *PostDeleter) DeleteThread(from string, account *postbar.Account, r *DeleteThreadRequest, withBan bool, givedBanReason string) {
	if withBan {
		if r.pid == 0 {
			r.pid = GetPidFromTid(r.tid, d.AccWin8)
		}

		banIDRequest := &BanIDRequest{
			tid:          r.tid,
			pid:          r.pid,
			spid:         0,
			uid:          r.uid,
			day:          1,
			userName:     r.author,
			title:        r.title,
			content:      r.content,
			postTime:     r.postTime,
			loggedReason: r.reason,
			givedReason:  givedBanReason,
		}

		if r.pid == 0 {
			d.OpLogger.Record(from, time.Now(), r, "无法获取主题pid,无法进行封禁,将不进行封禁.")
		} else {
			d.BanID(from, account.BDUSS, banIDRequest)
		}

	}

	if account.BDUSS == "" {
		d.OpLogger.Record(from, time.Now(), r, "忽略 未设置BDUSS")
		return
	}

	var status string
	now := time.Now()

	for i := 0; ; i++ {
		err, pberr := apis.DeleteThread(account, r.tid)
		if err == nil && (pberr == nil || pberr.ErrorCode == 0) {
			break
		} else if i < 3 {
			d.Logger.Error("删主题失败,将最多尝试三次:", err, pberr, ".")
		} else {
			d.Logger.Error("删主题失败,放弃:", err, pberr, ".")
			if err != nil {
				status = fmt.Sprint("失败", err)
			} else {
				status = fmt.Sprint("失败", pberr)
			}
			break
		}
	}

	d.OpLogger.Record(from, now, r, status)

}

type BanIDRequest struct {
	tid, pid, spid, uid uint64

	day      int
	userName string
	title    string
	content  interface{}
	postTime string

	loggedReason,
	givedReason,
	remark string
}

func (d *PostDeleter) BanID(from string, BDUSS string, r *BanIDRequest) {

	if BDUSS == "" {
		d.OpLogger.Record(from, time.Now(), r, "忽略 未设置BDUSS")
		return
	}

	var op_pid uint64
	if r.spid != 0 {
		op_pid = r.spid
	} else {
		op_pid = r.pid
	}

	var status string
	now := time.Now()

	for i := 0; ; i++ {
		err, pberr := apis.BlockIDWeb(BDUSS, d.ForumID, r.userName, op_pid, r.day, r.givedReason)
		if err == nil && (pberr == nil || pberr.ErrorCode == 0) {
			break
		} else if i < 3 {
			d.Logger.Error("封禁失败,将最多尝试三次:", err, pberr, ".")
		} else {
			d.Logger.Error("封禁失败,放弃:", err, pberr, ".")
			if err != nil {
				status = fmt.Sprint("失败", err)
			} else {
				status = fmt.Sprint("失败", pberr)
			}
			break
		}
	}

	d.OpLogger.Record(from, now, r, status)

}
