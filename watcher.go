package main

import (
	"github.com/samuel/go-zookeeper/zk"
)

type Watcher struct {
	conn *zk.Conn
	path string
}

func NewWatcher(conn *zk.Conn, path string) *Watcher {
	return &Watcher{conn, path}
}

func (w *Watcher) Start() (chan string, chan error) {
	changes := make(chan string)
	errors := make(chan error)

	go func() {
		watchTree(w.conn, w.path, changes, errors)
	}()

	return changes, errors
}

func (w *Watcher) Stop() {
	// TODO
}

func watchValue(conn *zk.Conn, path string, changes chan string, errors chan error) {
	for {
		_, _, events, err := conn.GetW(path)
		if err != nil {
			errors <- err
			return
		}

		evt := <-events
		// TODO: zk node does not exist is normal
		if evt.Err != nil {
			errors <- evt.Err
			return
		}

		changes <- path
	}
}

func watchTree(conn *zk.Conn, path string, changes chan string, errors chan error) {
	// TODO: don't need value watches on level 0 (i.e. /zconfig/servers)
	// TODO: this can also build the config
	treeWatches := make(map[string]bool)
	valueWatches := make(map[string]bool)

	for {
		children, _, events, err := conn.ChildrenW(path)
		if err != nil {
			errors <- err
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
					watchTree(conn, childpath, changes, errors)
				}()
			}

			// watch values
			if !valueWatches[child] {
				valueWatches[child] = true
				go func() {
					defer delete(valueWatches, child)
					watchValue(conn, childpath, changes, errors)
				}()
			}
		}

		evt := <-events
		// TODO: zk node does not exist is expected
		if evt.Err != nil {
			errors <- evt.Err
			return
		}

		changes <- path
	}
}
