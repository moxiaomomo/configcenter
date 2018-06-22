package watch

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/micro/go-micro/client"
	proto "github.com/moxiaomomo/configcenter/proto"
	"golang.org/x/net/context"
)

const (
	WatchTopic = "micro.config.watch"
)

var (
	mtx      sync.RWMutex
	watchers = make(map[string][]*watcher)
)

type watcher struct {
	id   string
	exit chan bool
	next chan *proto.WatchResponse
}

func ToWatchId(name, path, version string) string {
	return fmt.Sprintf("%s#*$#%s#*$#", name, path)
}

func FromWatchId(wathId string) (name, path, version string) {
	s := strings.Split(wathId, "#*$#")
	if len(s) != 3 {
		return "", "", ""
	}
	return s[0], s[1], s[2]
}

func (w *watcher) Next() (*proto.WatchResponse, error) {
	select {
	case c := <-w.next:
		return c, nil
	case <-w.exit:
		return nil, errors.New("watcher stopped")
	}
}

func (w *watcher) Stop() error {
	select {
	case <-w.exit:
		return errors.New("already stopped")
	default:
		close(w.exit)
	}

	mtx.Lock()
	defer mtx.Unlock()

	var wslice []*watcher

	for _, watch := range watchers[w.id] {
		if watch != w {
			wslice = append(wslice, watch)
		}
	}
	watchers[w.id] = wslice

	return nil
}

// function Watch created by a client RPC request
func Watch(id string) (*watcher, error) {
	mtx.Lock()
	defer mtx.Unlock()

	w := &watcher{
		id:   id,
		exit: make(chan bool),
		next: make(chan *proto.WatchResponse),
	}
	watchers[id] = append(watchers[id], w)
	return w, nil
}

// Used as a subscriber between config services for events
func Watcher(ctx context.Context, ch *proto.WatchResponse) error {
	mtx.Lock()
	defer mtx.Unlock()

	wathId := ToWatchId(ch.Name, ch.Path, ch.Version)
	for _, sub := range watchers[wathId] {
		select {
		case sub.next <- ch:
		case <-time.After(time.Millisecond * 100):
		}
	}
	return nil
}

func Publish(ctx context.Context, wresp *proto.WatchResponse) error {
	req := client.NewPublication(WatchTopic, wresp)
	return client.Publish(ctx, req)
}
