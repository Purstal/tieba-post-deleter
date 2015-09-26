package kw_manager

import (
	//"sync"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/purstal/go-tieba-base/logs"

	"github.com/purstal/go-tieba-modules/utils/action"
)

type StringKeywordManager struct {
	Logger *logs.Logger

	FileName      string
	StringSet     map[string]struct{}
	LastModTime   time.Time
	CheckInterval time.Duration

	actChan chan action.Action
}

func NewStringKeywordManager(logger *logs.Logger) *StringKeywordManager {
	return &StringKeywordManager{Logger: logger, actChan: make(chan action.Action),
		StringSet: map[string]struct{}{}}
}

func NewStringKeywordManagerBidingWithFile(keyWordFileFlieName string,
	checkInterval time.Duration, logger *logs.Logger) (*StringKeywordManager, error) {
	var m StringKeywordManager
	m.FileName = keyWordFileFlieName

	m.StringSet = map[string]struct{}{}

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
		err := LoadStrings(file, &m.StringSet, logger)
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
					err := LoadStrings(file, &m.StringSet, logger)
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

func (m StringKeywordManager) ChangeCheckInterval(newInterval time.Duration) {
	m.actChan <- action.Action{ChangeInterval, newInterval}
}

func (m StringKeywordManager) ChangeKeyWordFile(newFile string) {
	m.actChan <- action.Action{ChangeFile, newFile}
}

func (m StringKeywordManager) KeyWords() map[string]struct{} {
	return m.StringSet
}

func LoadStrings(file *os.File, set *map[string]struct{}, logger *logs.Logger) error {
	bytes, err := ReadAll(file)
	if err != nil {
		return err
	}

	lines := split([]rune(string(bytes)), []rune{'\n', ' '})

	newSet := map[string]struct{}{}

	oldSet := map[string]struct{}{}

	for k, v := range *set {
		oldSet[k] = v //复制一遍
	}

	var added []string

	for _, _line := range lines {
		var line string
		if line = strings.TrimSpace(string(_line)); line == "" {
			continue
		}
		newSet[line] = struct{}{}
		if _, exist := oldSet[line]; !exist {
			added = append(added, line)
		}
		delete(oldSet, line)
	}

	*set = newSet

	var updateInfo string = fmt.Sprintf("更新关键词(%s):", file.Name())
	if len(added) > 0 {
		updateInfo = updateInfo + "\n[+]"
		for _, str := range added {
			updateInfo = updateInfo + str + " "
		}
		updateInfo = strings.TrimRight(updateInfo, " ")
	}
	if len(oldSet) > 0 {
		updateInfo = updateInfo + "\n[-]"
		for str, _ := range oldSet {
			updateInfo = updateInfo + str + " "
		}
		updateInfo = strings.TrimRight(updateInfo, " ")
	}
	logger.Info(updateInfo)

	return nil
}

func split(s, seps []rune) [][]rune {
	var result [][]rune
	var last int
	for i, r := range s {
		for _, sep := range seps {
			if r == sep {
				if last != i {
					result = append(result, s[last:i])
				}
				last = i + 1
				continue
			}
		}
	}
	return result
}
