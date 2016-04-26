package main

import (
	"github.com/samuel/go-zookeeper/zk"
	"sync"
)

type Watcher struct {
	conn *zk.Conn
	path string

	changes chan string
	errors  chan error

	tree  map[string]struct{}
	value map[string]struct{}

	mutex *sync.Mutex
}

type WatchMethod int

const (
	WatchTree WatchMethod = iota
	WatchValue
)

func NewWatcher(conn *zk.Conn, path string) *Watcher {
	return &Watcher{
		conn:    conn,
		path:    path,
		changes: make(chan string),
		errors:  make(chan error),
		tree:    make(map[string]struct{}),
		value:   make(map[string]struct{}),
		mutex:   &sync.Mutex{},
	}
}

func (w *Watcher) Start() (chan string, chan error) {
	go func() {
		w.watchTree(w.path)
	}()

	return w.changes, w.errors
}

func (w *Watcher) Stop() {
	// TODO: signal to terminate zk watches
	// TODO: clear tree/value maps
	close(w.changes)
	close(w.errors)
}

func (w *Watcher) watchesFor(method WatchMethod) map[string]struct{} {
	switch method {
	case WatchTree:
		return w.tree
	case WatchValue:
		return w.value
	default:
		return nil
	}
}

func (w *Watcher) isWatching(method WatchMethod, path string) bool {
	watches := w.watchesFor(method)
	w.mutex.Lock()
	_, ok := watches[path]
	w.mutex.Unlock()
	return ok
}

func (w *Watcher) setWatching(method WatchMethod, path string, watching bool) {
	watches := w.watchesFor(method)
	w.mutex.Lock()
	if watching {
		watches[path] = struct{}{}
	} else {
		delete(watches, path)
	}
	w.mutex.Unlock()
}

func (w *Watcher) watchTree(path string) {
	if w.isWatching(WatchTree, path) {
		return
	}

	w.setWatching(WatchTree, path, true)
	defer w.setWatching(WatchTree, path, false)

	// TODO: don't need value watches on level 0 (i.e. /zconfig/servers)
	for {
		children, _, events, err := w.conn.ChildrenW(path)
		if err != nil {
			w.errors <- err
			return
		}

		for i := range children {
			child := children[i]
			childpath := path + "/" + child
			go w.watchTree(childpath)
			go w.watchValue(childpath)
		}

		evt := <-events
		if evt.Err != nil {
			w.errors <- evt.Err
			return
		}

		w.changes <- path
	}
}

func (w *Watcher) watchValue(path string) {
	if w.isWatching(WatchValue, path) {
		return
	}

	w.setWatching(WatchValue, path, true)
	defer w.setWatching(WatchValue, path, false)

	for {
		_, _, events, err := w.conn.GetW(path)
		if err != nil {
			w.errors <- err
			return
		}

		evt := <-events
		if evt.Err != nil {
			w.errors <- evt.Err
			return
		}

		w.changes <- path
	}
}
