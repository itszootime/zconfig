package main

import (
	"flag"
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"strings"
	"time"
)

var setup = NewSetup()

func init() {
	// TODO: check setup, ie. is the base path writable?
	flag.StringVar(&setup.BasePath, "base-path", "", "base path")
	flag.StringVar(&setup.Zk, "zk", "127.0.0.1:2181", "zk")
	flag.StringVar(&setup.ZkRoot, "zk-root", "/zconfig", "zk root")
	flag.Parse()
}

func iferr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	conn := zkConnect()
	defer conn.Close()
	zkInit(conn)

	cm := NewConfigManager(conn, setup.ZkRoot, setup.BasePath)
	err := cm.UpdateLocal()
	iferr(err)

	// get and watch root
	watcher := NewWatcher(conn, setup.ZkRoot)
	changes, errors := watcher.Start()
	for {
		select {
		case change := <-changes:
			fmt.Printf("change path=%v\n", change)
			err = cm.UpdateLocal()
			iferr(err)
		case <-errors:
			// we'll end up with node does not exist here
			// which will kill the go routine (it's fine)
			// fmt.Printf("main:error error=%v\n", err)
		}
	}
}

func zkConnect() *zk.Conn {
	// TODO: allow configurable timeout?
	conn, _, err := zk.Connect(strings.Split(setup.Zk, ","), time.Second)
	iferr(err) // severe
	return conn
}

func zkInit(conn *zk.Conn) {
	// TODO: ensure these flags are correct
	flags := int32(0)
	acl := zk.WorldACL(zk.PermAll)

	exists, _, err := conn.Exists(setup.ZkRoot)
	iferr(err) // severe
	if !exists {
		// TODO: ignore node already exists here
		_, err := conn.Create(setup.ZkRoot, nil, flags, acl)
		iferr(err)
	}
}
