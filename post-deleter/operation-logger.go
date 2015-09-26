package post_deleter

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/purstal/go-tieba-base/logs"
	"github.com/purstal/go-tieba-tools/operation-analyser/csv"
)

type OperationLogger struct {
	writer *csv.Writer
	lock   sync.Mutex
}

func NewOperationLogger(logDir string) (*OperationLogger, error) {
	var logger OperationLogger
	logFile, err := os.Create(logDir + "post-deleter-操作记录.csv")
	if err != nil {
		logs.Fatal("无法创建log文件.", err)
		return nil, err
	}

	logFile.Write([]byte{0xEF, 0xBB, 0xBF})

	logger.writer = csv.NewWriter(logFile)

	logger.writer.Write([]string{"操作", "状态", "操作来源", "原因", "本地操作时间", "发贴时间",
		"发贴人", "uid", "主题标题", "tid(主题id)", "pid(楼层id/楼中楼id)", "spid(楼中楼id)", "贴子内容", "其他"})

	return &logger, nil
}

func (l *OperationLogger) Record(from string, _opTime time.Time, _r interface{}, status string) {
	opTime := _opTime.Format("2006-01-02 15:04:05")
	switch _r.(type) {
	case *DeletePostRequest:
		r := _r.(*DeletePostRequest)
		logs.Info(MakePrefix(nil, r.tid, r.pid, r.spid, r.uid), from, "删贴", r.reason, status)
		uid, tid, pid, spid := strconv.FormatUint(r.uid, 10), strconv.FormatUint(r.tid, 10),
			strconv.FormatUint(r.pid, 10), strconv.FormatUint(r.spid, 10)
		l.writer.Write([]string{"删贴", status, from, r.reason, opTime, r.postTime,
			r.author, uid, r.title, tid, pid, spid,
			string(replace([]rune(fmt.Sprint(r.content)), '\n', '\t')), r.remark})
	case *DeleteThreadRequest:
		r := _r.(*DeleteThreadRequest)
		logs.Info(MakePrefix(nil, r.tid, r.pid, 0, r.uid), from, "删主题", r.reason, status)
		uid, tid, pid := strconv.FormatUint(r.uid, 10), strconv.FormatUint(r.tid, 10),
			strconv.FormatUint(r.pid, 10)
		l.writer.Write([]string{"删主题", status, from, r.reason, opTime, r.postTime,
			r.author, uid, r.title, tid, pid, "0",
			string(replace([]rune(fmt.Sprint(r.content)), '\n', '\t')), r.remark})
	case *BanIDRequest:
		r := _r.(*BanIDRequest)
		logs.Info(MakePrefix(nil, r.tid, r.pid, r.spid, r.uid), from, "封禁", fmt.Sprintf("%s(%s)", r.loggedReason, r.givedReason), status)
		tid, pid, spid, uid := strconv.FormatUint(r.tid, 10),
			strconv.FormatUint(r.pid, 10), strconv.FormatUint(r.spid, 10), strconv.FormatUint(r.uid, 10)
		l.writer.Write([]string{"封禁", status, from, r.loggedReason, opTime, r.postTime,
			r.userName, uid, r.title, tid, pid, spid,
			string(replace([]rune(fmt.Sprint(r.content)), '\n', '\t')), r.remark + fmt.Sprintf("天数:%d, 给出原因:%s, ", r.day, r.givedReason)})
	}

	l.writer.Flush()
}

func replace(src []rune, a, b rune) []rune {

	dst := make([]rune, len(src))

	for i, r := range src {
		if r == a {
			dst[i] = b
		} else {
			dst[i] = r
		}
	}

	return dst
}
