package main

import (
	"github.com/samuel/go-zookeeper/zk"
)

type Watcher struct {
	conn *zk.Conn
	path string

	changes chan string
	errors  chan error
}

func NewWatcher(conn *zk.Conn, path string) *Watcher {
	return &Watcher{conn, path, make(chan string), make(chan error)}
}

func (w *Watcher) Start() (chan string, chan error) {
	go func() {
		w.watchTree(w.path)
	}()

	return w.changes, w.errors
}

func (w *Watcher) Stop() {
	// TODO
}

func (w *Watcher) watchValue(path string) {
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

func (w *Watcher) watchTree(path string) {
	// TODO: don't need value watches on level 0 (i.e. /zconfig/servers)
	// TODO: this can also build the config
	treeWatches := make(map[string]bool)
	valueWatches := make(map[string]bool)

	for {
		children, _, events, err := w.conn.ChildrenW(path)
		if err != nil {
			w.errors <- err
			return
		}

		for i := range children {
			child := children[i]
			childpath := path + "/" + child

			// watch tree
			if !treeWatches[child] {
				treeWatches[child] = true
				go func() {
					defer delete(treeWatches, child)
					w.watchTree(childpath)
				}()
			}

			// watch values
			if !valueWatches[child] {
				valueWatches[child] = true
				go func() {
					defer delete(valueWatches, child)
					w.watchValue(childpath)
				}()
			}
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
