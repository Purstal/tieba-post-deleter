package post_deleter

import (
	"strings"
	"time"

	"github.com/purstal/go-tieba-base/tieba"

	postfinder "github.com/purstal/go-tieba-modules/post-finder"

	"github.com/purstal/tieba-post-deleter/post-deleter/keyword-manager"
)

func MakePrefix(serverTime *time.Time, tid, pid, spid, uid uint64) string {
	return postfinder.MakePostLogString(serverTime, tid, pid, spid, uid)
}

func GetServerTimeFromExtra(extra postbar.IExtra) *time.Time {
	return postfinder.GetServerTimeFromExtra(extra)

}

func ExtractText(contentList []postbar.Content) string {
	var str string
	for _, content := range contentList {
		if text, ok := content.(postbar.Text); ok {
			str = str + text.Text + "\n"
		}
	}
	return strings.TrimSuffix(str, "\n")
}

func SliceContainsString(slice []string, sub string) bool {
	for _, str := range slice {
		if sub == str {
			return true
		}
	}
	return false
}

func MatchAny(text string, exps []kw_manager.RegexpKeyword) *kw_manager.RegexpKeyword {
	for _, exp := range exps {
		if exp.Rx.MatchString(text) {
			return &exp
		}
	}
	return nil
}

func InStringSet(set map[string]struct{}, key string) bool {
	_, exist := set[key]
	return exist
}
