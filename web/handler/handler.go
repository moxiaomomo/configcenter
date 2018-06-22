package handler

import (
	"crypto/sha1"
	"fmt"
	"html/template"
	"net/http"
	// "net/url"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/yosssi/ace"
	"golang.org/x/net/context"

	"github.com/moxiaomomo/configcenter/logger"
	proto "github.com/moxiaomomo/configcenter/proto"
)

var (
	opts         *ace.Options
	configClient proto.ConfigClient
)

const (
	AUDITLOG_PAGE_LIMIT = 15
)

func Init(dir string, t proto.ConfigClient) {
	configClient = t

	opts = ace.InitializeOptions(nil)
	opts.BaseDir = dir
	opts.DynamicReload = true
	opts.FuncMap = template.FuncMap{
		"JSON": func(d string) string {
			return prettyJSON(d)
		},
		"TimeAgo": func(t int64) string {
			return timeAgo(t)
		},
		"TimeStamp": func(t int64) string {
			return time.Unix(t, 0).Format("2006-01-02 15:04:05")
		},
		"Colour": func(s string) string {
			return colour(s)
		},
		"ConfStatus": func(s int32) string {
			return confStatusStr(s)
		},
		"ConfColor": func(s int32) string {
			return confStatusColor(s)
		},
		"Int32ToStr": func(s int32) string {
			return fmt.Sprintf("%d", s)
		},
	}
}

func render(w http.ResponseWriter, r *http.Request, tmpl string, data map[string]interface{}) {
	basePath := hostPath(r)

	opts.FuncMap["URL"] = func(path string) string {
		return filepath.Join(basePath, path)
	}

	tpl, err := ace.Load("layout", tmpl, opts)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "r", 302)
		return
	}

	if data == nil {
		data = make(map[string]interface{})
	}
	data["Alert"] = getAlert(w, r)

	if err := tpl.Execute(w, data); err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/", 302)
	}
}

func Index(w http.ResponseWriter, r *http.Request) {
	rsp, err := configClient.AuditLog(context.TODO(), &proto.AuditLogRequest{
		Limit:   int64(10),
		Reverse: true,
	})
	if err != nil {
		http.Redirect(w, r, "/", 302)
		return
	}

	sort.Sort(sortedLogs{logs: rsp.Changes, reverse: false})
	render(w, r, "index", map[string]interface{}{
		"Latest": rsp.Changes,
	})
}

// The Audit Log
func AuditLog(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	limit := AUDITLOG_PAGE_LIMIT
	from, _ := strconv.Atoi(r.Form.Get("from"))
	to, _ := strconv.Atoi(r.Form.Get("to"))

	page, err := strconv.Atoi(r.Form.Get("page"))
	if err != nil {
		page = 1
	}

	if page < 1 {
		page = 1
	}

	offset := (page * limit) - limit

	rsp, err := configClient.AuditLog(context.TODO(), &proto.AuditLogRequest{
		From:    int64(from),
		To:      int64(to),
		Limit:   int64(limit),
		Offset:  int64(offset),
		Reverse: true,
	})
	if err != nil {
		http.Redirect(w, r, "/", 302)
		return
	}

	sort.Sort(sortedLogs{logs: rsp.Changes, reverse: false})

	var less, more int

	if len(rsp.Changes) == limit {
		more = page + 1
	}

	if page > 1 {
		less = page - 1
	}

	render(w, r, "audit", map[string]interface{}{
		"Latest": rsp.Changes,
		"Less":   less,
		"More":   more,
	})
}

func ConfigCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		render(w, r, "cfgcreateidx", map[string]interface{}{
			"CreateStatus": map[string]string{
				"0": ConfStatus["0"],
				"1": ConfStatus["1"],
			},
		})
		return
	}

	r.ParseForm()
	name := r.Form.Get("name")
	path := r.Form.Get("path")
	version := r.Form.Get("version")
	_status := r.Form.Get("status")
	comment := r.Form.Get("comment")
	conf := r.Form.Get("config")

	if len(name) == 0 || len(path) == 0 || len(version) == 0 {
		setAlert(w, r, "Name/Path/Version不能为空", "error")
		http.Redirect(w, r, r.Referer(), 302)
		return
	}

	status, err := strconv.ParseInt(_status, 10, 32)
	if err != nil {
		setAlert(w, r, "Status值无效", "error")
		http.Redirect(w, r, r.Referer(), 302)
		return
	}

	sum := fmt.Sprintf("%x", sha1.Sum([]byte(conf)))

	_, err = configClient.Create(context.TODO(), &proto.CreateRequest{
		ConfigSet: &proto.ConfigSet{
			Name:    name,
			Path:    path,
			Version: version,
			Comment: comment,
			Status:  int32(status),
			ChangeSet: &proto.ChangeSet{
				Timestamp: time.Now().Unix(),
				Checksum:  sum,
				Data:      conf,
				Source:    "web",
			},
		},
	})
	if err != nil {
		setAlert(w, r, err.Error(), "error")
		http.Redirect(w, r, r.Referer(), 302)
		return
	}
	url := fmt.Sprintf("/config/query?name=%s&path=%s&version=%s&up=0", name, path, version)
	http.Redirect(w, r, url, 302)
	return
}

func ConfigIndx(w http.ResponseWriter, r *http.Request) {
	render(w, r, "configindx", nil)
}

func ConfigQuery(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if r.Method == "GET" {
		if r.Form.Get("up") == "1" {
			setAlert(w, r, "配置更改成功", "info")
		} else if r.Form.Get("up") == "0" {
			setAlert(w, r, "配置创建成功", "info")
		}
	}

	name := r.Form.Get("name")
	path := r.Form.Get("path")
	version := r.Form.Get("version")
	// logger.Infof("name:%s path:%s version:%s\n", name, path, version)

	if len(name) == 0 || len(path) == 0 || len(version) == 0 {
		setAlert(w, r, "Name/Path/Version不能为空", "error")
		http.Redirect(w, r, r.Referer(), 302)
		return
	}

	rsp, err := configClient.Read(context.TODO(), &proto.ReadRequest{
		Name:       name,
		Path:       path,
		Version:    version,
		DelInclude: true,
	})

	if err != nil {
		setAlert(w, r, err.Error(), "error")
		http.Redirect(w, r, r.Referer(), 302)
		return
	} else if len(name) == 0 {
		setAlert(w, r, "没有找到相关的配置项", "error")
		http.Redirect(w, r, r.Referer(), 302)
		return
	}

	render(w, r, "configquery", map[string]interface{}{
		"Name":      name,
		"Path":      path,
		"Version":   version,
		"ConfigSet": []*proto.ConfigSet{rsp.ConfigSet},
	})
	return
}

func ConfigEdit(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	name := r.Form.Get("name")
	path := r.Form.Get("path")
	version := r.Form.Get("version")
	// logger.Infof("name:%s path:%s ver:%s\n", name, path, version)

	rsp, err := configClient.Read(context.TODO(), &proto.ReadRequest{
		Name:    name,
		Path:    path,
		Version: version,
	})

	if err != nil {
		setAlert(w, r, err.Error(), "error")
		http.Redirect(w, r, r.Referer(), 302)
		return
	} else if len(name) == 0 {
		setAlert(w, r, "没有找到相关的配置项", "error")
		http.Redirect(w, r, r.Referer(), 302)
		return
	}

	render(w, r, "configedit", map[string]interface{}{
		"Name":       name,
		"Path":       path,
		"Version":    version,
		"Status":     fmt.Sprintf("%d", rsp.ConfigSet.Status),
		"ConfStatus": ConfStatus,
		"ConfigSet":  []*proto.ConfigSet{rsp.ConfigSet},
	})
}

func ConfigUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		setAlert(w, r, "Unsupported method.", "error")
		http.Redirect(w, r, r.Referer(), 302)
		return
	}

	r.ParseForm()
	name := r.Form.Get("name")
	path := r.Form.Get("path")
	version := r.Form.Get("version")
	_status := r.Form.Get("status")
	comment := r.Form.Get("comment")
	conf := r.Form.Get("config")

	if len(name) == 0 || len(path) == 0 || len(version) == 0 {
		setAlert(w, r, "Name/Path/Version不能为空", "error")
		http.Redirect(w, r, r.Referer(), 302)
		return
	}

	status, err := strconv.ParseInt(_status, 10, 32)
	if err != nil {
		setAlert(w, r, "Status值无效", "error")
		http.Redirect(w, r, r.Referer(), 302)
		return
	}

	sum := fmt.Sprintf("%x", sha1.Sum([]byte(conf)))
	_, err = configClient.Update(context.TODO(), &proto.UpdateRequest{
		ConfigSet: &proto.ConfigSet{
			Name:    name,
			Path:    path,
			Version: version,
			Status:  int32(status),
			Comment: comment,
			ChangeSet: &proto.ChangeSet{
				Timestamp: time.Now().Unix(),
				Checksum:  sum,
				Data:      conf,
				Source:    "web",
			},
		},
	})
	if err != nil {
		setAlert(w, r, err.Error(), "error")
		//		http.Redirect(w, r, r.Referer(), 302)
		return
	}

	url := fmt.Sprintf("/config/query?name=%s&path=%s&version=%s&up=1", name, path, version)
	logger.Infof("qurl: %s\n", url)
	http.Redirect(w, r, url, 302)
	return
}

func ConfigSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		render(w, r, "cfgsearchidx", nil)
		return
	}

	r.ParseForm()
	limit := 25
	name := r.Form.Get("name")
	path := r.Form.Get("path")

	if len(name) == 0 {
		setAlert(w, r, "Name不能为空", "error")
		http.Redirect(w, r, r.Referer(), 302)
		return
	}

	page, err := strconv.Atoi(r.Form.Get("p"))
	if err != nil {
		page = 1
	}

	if page < 1 {
		page = 1
	}

	offset := (page * limit) - limit

	rsp, err := configClient.Search(context.TODO(), &proto.SearchRequest{
		Name:   name,
		Path:   path,
		Limit:  int64(limit),
		Offset: int64(offset),
	})
	if err != nil {
		http.Redirect(w, r, filepath.Join(hostPath(r), "search"), 302)
		return
	}

	q := ""

	if len(name) > 0 {
		q += "name: " + name + ", "
	}

	if len(path) > 0 {
		q += "path: " + path
	}

	var less, more int

	if len(rsp.Configs) == limit {
		more = page + 1
	}

	if page > 1 {
		less = page - 1
	}

	sort.Sort(sortedConfigs{configs: rsp.Configs})

	render(w, r, "cfgsearchres", map[string]interface{}{
		"Name":    name,
		"Path":    path,
		"Results": rsp.Configs,
		"Less":    less,
		"More":    more,
	})
}
