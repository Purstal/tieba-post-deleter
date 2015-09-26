package kw_manager

import (
	//"sync"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/purstal/go-tieba-base/logs"

	"github.com/purstal/go-tieba-modules/utils/action"
)

type Uint64KeywordManager struct {
	Logger *logs.Logger

	FileName      string
	Uint64Set     map[uint64]struct{}
	LastModTime   time.Time
	CheckInterval time.Duration

	actChan chan action.Action
}

func NewUint64KeywordManager(logger *logs.Logger) *Uint64KeywordManager {
	return &Uint64KeywordManager{Logger: logger, actChan: make(chan action.Action),
		Uint64Set: map[uint64]struct{}{}}
}

func NewUint64KeywordManagerBidingWithFile(keyWordFileFlieName string,
	checkInterval time.Duration, logger *logs.Logger) (*Uint64KeywordManager, error) {
	var m Uint64KeywordManager
	m.FileName = keyWordFileFlieName

	m.Uint64Set = map[uint64]struct{}{}

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
		err := LoadUint64s(file, &m.Uint64Set, logger)
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
					err := LoadUint64s(file, &m.Uint64Set, logger)
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

func (m Uint64KeywordManager) ChangeCheckInterval(newInterval time.Duration) {
	m.actChan <- action.Action{ChangeInterval, newInterval}
}

func (m Uint64KeywordManager) ChangeKeyWordFile(newFile string) {
	m.actChan <- action.Action{ChangeFile, newFile}
}

func (m Uint64KeywordManager) KeyWords() map[uint64]struct{} {
	return m.Uint64Set
}

func LoadUint64s(file *os.File, set *map[uint64]struct{}, logger *logs.Logger) error {
	bytes, err := ReadAll(file)
	if err != nil {
		return err
	}

	lines := split([]rune(string(bytes)), []rune{'\n', ' '})

	newSet := map[uint64]struct{}{}
	oldSet := map[uint64]struct{}{}

	for k, v := range *set {
		oldSet[k] = v //复制一遍
	}

	var added []uint64

	for lineNo, _line := range lines {
		var line string
		if line = strings.TrimSpace(string(_line)); line == "" {
			continue
		}
		u64, err := strconv.ParseUint(line, 10, 64)
		if err != nil {
			logs.Error(fmt.Sprintf("不正确的关键词(第%d行),跳过.", lineNo), err)
		}
		newSet[u64] = struct{}{}
		if _, exist := oldSet[u64]; !exist {
			added = append(added, u64)
		}
		delete(oldSet, u64)
	}

	*set = newSet

	var updateInfo string = fmt.Sprintf("更新关键词(%s):", file.Name())
	if len(added) > 0 {
		updateInfo = updateInfo + "\n[+]"
		for _, u64 := range added {
			updateInfo = updateInfo + strconv.FormatUint(u64, 10) + " "
		}
		updateInfo = strings.TrimRight(updateInfo, " ")
	}

	if len(oldSet) > 0 {
		updateInfo = updateInfo + "\n[-]"
		for u64, _ := range oldSet {
			updateInfo = updateInfo + strconv.FormatUint(u64, 10) + " "
		}
		updateInfo = strings.TrimRight(updateInfo, " ")
	}

	logger.Info(updateInfo)

	return nil
}
