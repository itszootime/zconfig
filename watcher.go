package main

import (
	"github.com/samuel/go-zookeeper/zk"
)

type Watcher struct {
	conn *zk.Conn
	path string

	changes chan string
	errors  chan error

	tree  map[string]bool
	value map[string]bool
}

func NewWatcher(conn *zk.Conn, path string) *Watcher {
	return &Watcher{
		conn:    conn,
		path:    path,
		changes: make(chan string),
		errors:  make(chan error),
		tree:    make(map[string]bool),
		value:   make(map[string]bool),
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

func (w *Watcher) watchTree(path string) {
	// TODO: mutex here
	if _, ok := w.tree[path]; ok {
		return
	}

	w.tree[path] = true
	defer delete(w.tree, path)

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
		// TODO: zk node does not exist is expected
		if evt.Err != nil {
			w.errors <- evt.Err
			return
		}

		w.changes <- path
	}
}

func (w *Watcher) watchValue(path string) {
	// TODO: mutex here
	if _, ok := w.value[path]; ok {
		return
	}

	w.value[path] = true
	defer delete(w.value, path)

	for {
		_, _, events, err := w.conn.GetW(path)
		if err != nil {
			w.errors <- err
			return
		}

		evt := <-events
		// TODO: zk node does not exist is normal
		if evt.Err != nil {
			w.errors <- evt.Err
			return
		}

		w.changes <- path
	}
}
