package post_deleter

import (
	"os"
	"time"

	//"github.com/mitchellh/go-mruby"

	"github.com/purstal/go-tieba-base/logs"
	"github.com/purstal/go-tieba-base/tieba"

	postfinder "github.com/purstal/go-tieba-modules/post-finder"

	"github.com/purstal/go-tieba-tools/simple-post-deleter/post-deleter/keyword-manager"
)

type PostDeleter struct {
	AccWin8, AccAndr *postbar.Account
	PostFinder       *postfinder.PostFinder

	Content_RxKw,
	UserName_RxKw *kw_manager.RegexpKeywordManager

	Tid_Whitelist *kw_manager.Uint64KeywordManager

	UserName_Whitelist,
	BawuList *kw_manager.StringKeywordManager

	ForumName string
	ForumID   uint64

	Records Records

	Logger   *logs.Logger
	OpLogger *OperationLogger
}

type PostDeleterBuildingParameters struct {
	AccWin8, AccAndr *postbar.Account

	ForumName string
	ForumID   uint64

	ConfgiFileName ConfgiFileName

	//Mrb *mruby.Mrb

	Debugging bool
	LogDir    string
}

type ConfgiFileName struct {
	ContentRegexp,
	UserNameRegexp,
	TidWhiteList,
	UserNameWhiteList,
	BawuList string
}

func NewPostDeleter(b PostDeleterBuildingParameters) (*PostDeleter, error) {
	var deleter PostDeleter

	if err := initLogger(&deleter, b.LogDir); err != nil {
		return nil, err
	}

	if opLogger, err := NewOperationLogger(b.LogDir); err != nil {
		return nil, err
	} else {
		deleter.OpLogger = opLogger
	}

	{
		var err error
		if deleter.Content_RxKw, err = newRxKwManager(b.ConfgiFileName.ContentRegexp, deleter.Logger); err != nil {
			return nil, err
		} else if deleter.UserName_RxKw, err = newRxKwManager(b.ConfgiFileName.UserNameRegexp, deleter.Logger); err != nil {
			return nil, err
		} else if deleter.Tid_Whitelist, err = newU64KwManager(b.ConfgiFileName.TidWhiteList, deleter.Logger); err != nil {
			return nil, err
		} else if deleter.UserName_Whitelist, err = newStrKwManager(b.ConfgiFileName.UserNameWhiteList, deleter.Logger); err != nil {
			return nil, err
		} else if deleter.BawuList, err = newStrKwManager(b.ConfgiFileName.BawuList, deleter.Logger); err != nil {
			return nil, err
		}
	}

	deleter.AccWin8, deleter.AccAndr = b.AccWin8, b.AccAndr
	deleter.ForumID, deleter.ForumName = b.ForumID, b.ForumName

	if postFinder, err := postfinder.NewPostFinder(
		deleter.AccWin8, deleter.AccAndr, deleter.ForumName,
		func(postfinder *postfinder.PostFinder) {
			postfinder.ThreadFilter = deleter.ThreadFilter
			postfinder.NewThreadFirstAssessor = deleter.NewThreadFirstAssessor
			postfinder.NewThreadSecondAssessor = deleter.NewThreadSecondAssessor
			postfinder.AdvSearchAssessor = deleter.AdvSearchAssessor
			postfinder.PostAssessor = deleter.PostAssessor
			postfinder.CommentAssessor = deleter.CommentAssessor
		}, b.Debugging, b.LogDir); err != nil {
		return nil, err
	} else {
		deleter.PostFinder = postFinder
	}

	deleter.Records.RulesThread_Tids, deleter.Records.ServerListThread_Tids,
		deleter.Records.WaterThread_Tids =
		map[uint64]struct{}{}, map[uint64]struct{}{}, map[uint64]struct{}{}

	return &deleter, nil
}

func (p *PostDeleter) Run(monitorInterval time.Duration) {
	p.PostFinder.Run(monitorInterval)
}

func initLogger(pd *PostDeleter, logDir string) error {
	logFile, err := os.Create(logDir + "post-deleter-日志.log")
	if err != nil {
		logs.Fatal("无法创建log文件.", err)
		return err
	}
	pd.Logger = logs.NewLogger(logs.DebugLevel, os.Stdout, logFile)
	logs.DefaultLogger = pd.Logger
	return nil
}

func newRxKwManager(fileName string, logger *logs.Logger) (*kw_manager.RegexpKeywordManager, error) {
	var m *kw_manager.RegexpKeywordManager
	var err error
	if fileName != "" {
		m, err =
			kw_manager.NewRegexpKeywordManagerBidingWithFile(
				fileName, time.Second, logger)
		if err != nil {
			logger.Error("无法创建正则关键词管理.", err)
			return nil, err
		}
		return m, nil
	} else {
		logger.Warn("未设置正则关键词文件")
		return kw_manager.NewRegexpKeywordManager(logger), nil
	}
}

func newU64KwManager(fileName string, logger *logs.Logger) (*kw_manager.Uint64KeywordManager, error) {
	var m *kw_manager.Uint64KeywordManager
	var err error
	if fileName != "" {
		m, err =
			kw_manager.NewUint64KeywordManagerBidingWithFile(
				fileName, time.Second, logger)
		if err != nil {
			logger.Error("无法创建uint64关键词管理.", err)
			return nil, err
		}
		return m, nil
	} else {
		logger.Warn("未设置uint64关键词文件")
		return kw_manager.NewUint64KeywordManager(logger), nil
	}
}

func newStrKwManager(fileName string, logger *logs.Logger) (*kw_manager.StringKeywordManager, error) {
	var m *kw_manager.StringKeywordManager
	var err error
	if fileName != "" {
		m, err =
			kw_manager.NewStringKeywordManagerBidingWithFile(
				fileName, time.Second, logger)
		if err != nil {
			logger.Error("无法创建string关键词管理.", err)
			return nil, err
		}
		return m, nil
	} else {
		logger.Warn("未设置string关键词文件")
		return kw_manager.NewStringKeywordManager(logger), nil
	}
}
