package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	proto "github.com/moxiaomomo/configcenter/proto"
)

const (
	alertId = "_s"
)

var (
	colours    = []string{"blue", "green", "yellow", "purple", "orange"}
	store      = sessions.NewCookieStore([]byte("config"))
	ConfStatus = map[string]string{
		"0":  "Disabled",
		"1":  "Runable",
		"-1": "Deleted",
	}
)

type Alert struct {
	Type, Message string
}

type sortedConfigs struct {
	configs []*proto.ConfigSet
}

type sortedLogs struct {
	logs    []*proto.ChangeLog
	reverse bool
}

func (s sortedConfigs) Len() int {
	return len(s.configs)
}

func (s sortedConfigs) Less(i, j int) bool {
	return s.configs[i].ChangeSet.Timestamp > s.configs[j].ChangeSet.Timestamp
}

func (s sortedConfigs) Swap(i, j int) {
	s.configs[i], s.configs[j] = s.configs[j], s.configs[i]
}

func (s sortedLogs) Len() int {
	return len(s.logs)
}

func (s sortedLogs) Less(i, j int) bool {
	if s.reverse {
		return s.logs[i].ConfigSet.UpdatedAt < s.logs[j].ConfigSet.UpdatedAt
	}
	return s.logs[i].ConfigSet.UpdatedAt > s.logs[j].ConfigSet.UpdatedAt
}

func (s sortedLogs) Swap(i, j int) {
	s.logs[i], s.logs[j] = s.logs[j], s.logs[i]
}

func getAlert(w http.ResponseWriter, r *http.Request) *Alert {
	sess, err := store.Get(r, alertId)
	if err != nil {
		return nil
	}
	defer sess.Save(r, w)

	for _, i := range []string{"info", "error", "success"} {
		f := sess.Flashes(i)
		if f != nil {
			if i == "error" {
				i = "danger"
			}
			return &Alert{
				Type:    i,
				Message: f[0].(string),
			}
		}
	}
	return nil
}

func setAlert(w http.ResponseWriter, r *http.Request, msg string, typ string) {
	sess, err := store.Get(r, alertId)
	if err != nil {
		return
	}
	sess.AddFlash(msg, typ)
	sess.Save(r, w)
}

func colour(s string) string {
	return colours[crc32.ChecksumIEEE([]byte(s))%uint32(len(colours))]
}

func distanceOfTime(minutes float64) string {
	switch {
	case minutes < 1:
		return fmt.Sprintf("%d secs", int(minutes*60))
	case minutes < 59:
		return fmt.Sprintf("%d minutes", int(minutes))
	case minutes < 90:
		return "about an hour"
	case minutes < 120:
		return "almost 2 hours"
	case minutes < 1080:
		return fmt.Sprintf("%d hours", int(minutes/60))
	case minutes < 1680:
		return "about a day"
	case minutes < 2160:
		return "more than a day"
	case minutes < 2520:
		return "almost 2 days"
	case minutes < 2880:
		return "about 2 days"
	default:
		return fmt.Sprintf("%d days", int(minutes/1440))
	}
	return ""
}

func prettyJSON(d string) string {
	b := bytes.NewBuffer(nil)
	defer b.Reset()

	err := json.Indent(b, []byte(d), "", "\t")
	if err != nil {
		return d
	}
	return b.String()
}

func timeAgo(t int64) string {
	d := time.Unix(t, 0)
	timeAgo := ""
	startDate := time.Now().Unix()
	deltaMinutes := float64(startDate-d.Unix()) / 60.0
	if deltaMinutes <= 523440 {
		timeAgo = fmt.Sprintf("%s ago", distanceOfTime(deltaMinutes))
	} else {
		timeAgo = d.Format("2 Jan")
	}

	return timeAgo
}

func hostPath(r *http.Request) string {
	if path := r.Header.Get("x-Micro-Web-Base-Path"); len(path) > 0 {
		return path
	}
	return "/"
}

func confStatusStr(s int32) string {
	switch s {
	case -1:
		return "Deleted"
	case 0:
		return "Disabled"
	case 1:
		return "Runable"
	default:
		return "Unknown"
	}
}

func confStatusColor(s int32) string {
	switch s {
	case -1:
		return "red"
	case 0:
		return "orange"
	case 1:
		return "yellowgreen"
	case 2:
		return "green"
	default:
		return "black"
	}
}
