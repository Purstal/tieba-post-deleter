package kw_manager

import (
	//"sync"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/purstal/go-tieba-base/logs"

	"github.com/purstal/go-tieba-modules/utils/action"
)

const (
	ChangeInterval action.Pattern = iota
	ChangeFile
)

type RegexpKeyword struct {
	BanFlag bool
	Rx      *regexp.Regexp
}

type RegexpKeywordManager struct {
	Logger *logs.Logger

	FileName      string
	KewWordExps   []RegexpKeyword
	LastModTime   time.Time
	CheckInterval time.Duration

	actChan chan action.Action
}

func NewRegexpKeywordManager(logger *logs.Logger) *RegexpKeywordManager {
	return &RegexpKeywordManager{Logger: logger, actChan: make(chan action.Action)}
}

func NewRegexpKeywordManagerBidingWithFile(keyWordFileFlieName string,
	checkInterval time.Duration, logger *logs.Logger) (*RegexpKeywordManager, error) {
	var m RegexpKeywordManager
	m.FileName = keyWordFileFlieName

	file, err := os.Open(m.FileName)
	if os.IsNotExist(err) {
		var err error
		file, err = os.Create(m.FileName)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else {
		err := LoadExps(file, &m.KewWordExps, logger)
		if err != nil {
			return nil, err
		}
	}

	if fi, err := file.Stat(); err != nil {
		return nil, err
	} else {
		m.LastModTime = fi.ModTime()
	}
	file.Close()

	m.CheckInterval = checkInterval
	m.actChan = make(chan action.Action)

	go func() {
		ticker := time.NewTicker(m.CheckInterval)
		for {
			select {
			case <-ticker.C:
			case act := <-m.actChan:
				switch act.Pattern {
				case ChangeInterval:
					ticker.Stop()
					ticker = time.NewTicker(act.Param.(time.Duration))
				case ChangeFile:
				}
				continue
			}
			file, err1 := os.Open(m.FileName)
			if err1 != nil {
				logger.Error("无法打开关键词文件,跳过本次.", err1, ".")
				continue
			}
			func() {
				defer func() { file.Close() }()
				fi, err2 := file.Stat()
				if err2 != nil {
					logger.Error("无法获取文件信息,跳过本次.", err2, ".")
					return
				}
				if modTime := fi.ModTime(); modTime != m.LastModTime {
					err := LoadExps(file, &m.KewWordExps, logger)
					if err != nil {
						logger.Error("无法更新关键词,下次修改前将不尝试读取.", err, ".")
					}
					m.LastModTime = modTime
				}
			}()
		}
	}()

	return &m, nil
}

func (m RegexpKeywordManager) ChangeCheckInterval(newInterval time.Duration) {
	m.actChan <- action.Action{ChangeInterval, newInterval}
}

func (m RegexpKeywordManager) ChangeKeyWordFile(newFile string) {
	m.actChan <- action.Action{ChangeFile, newFile}
}

func (m RegexpKeywordManager) KeyWords() []RegexpKeyword {
	return m.KewWordExps
}

func LoadExps(file *os.File, exps *[]RegexpKeyword, logger *logs.Logger) error {

	bytes, err := ReadAll(file)
	if err != nil {
		return err
	}

	lines := strings.Split(string(bytes), "\n")

	oldExps := make(map[string]RegexpKeyword)

	for _, exp := range *exps {
		if exp.BanFlag {
			oldExps["$ban "+exp.Rx.String()] = exp
		} else {
			oldExps[exp.Rx.String()] = exp
		}
	}

	newExps := make(map[string]RegexpKeyword)
	var addedExps []string

	for lineNo, line := range lines {
		line = strings.TrimRightFunc(line, func(r rune) bool {
			return r == '\n' || r == '\r'
		})
		if line == "" {
			continue
		}
		if exp, exist := oldExps[line]; exist {
			newExps[line] = exp
			delete(oldExps, line)
		} else {
			var banFlag bool
			var newExp *regexp.Regexp
			var err error
			if banFlag = strings.HasPrefix(line, "$ban "); banFlag {
				newExp, err = regexp.Compile(strings.TrimLeft(line, "$ban "))
			} else {
				newExp, err = regexp.Compile(line)
			}
			if err != nil {
				logs.Error(fmt.Sprintf("不正确的关键词(第%d行),跳过.", lineNo), err)
			} else {
				newExps[line] = RegexpKeyword{banFlag, newExp}
				addedExps = append(addedExps, line)
			}

		}
	}

	newExpSlice := make([]RegexpKeyword, 0, len(newExps))

	for _, exp := range newExps {
		newExpSlice = append(newExpSlice, exp)
	}

	*exps = newExpSlice

	var updateInfo string = fmt.Sprintf("更新关键词(%s):\n", file.Name())
	for _, exp := range addedExps {
		updateInfo = updateInfo + "[+] " + exp + "\n"
	}
	for _, exp := range oldExps {
		if exp.BanFlag {
			updateInfo = updateInfo + "[-] $ban" + exp.Rx.String() + "\n"
		} else {
			updateInfo = updateInfo + "[-] " + exp.Rx.String() + "\n"
		}
	}
	updateInfo = strings.TrimSuffix(updateInfo, "\n")
	logger.Info(updateInfo)

	//logger.Debug("现在的关键词:", newExpSlice, ".")

	return nil
}

func ReadAll(file *os.File) ([]byte, error) {
	var data, err = ioutil.ReadAll(file)

	if err != nil {
		return data, err
	}

	if len(data) < 3 {
		return data, nil
	}

	if data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {

		return data[3:], nil
	}

	return data, nil
}
