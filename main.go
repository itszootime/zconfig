package main

import (
	"flag"
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"strings"
)

var setup = NewSetup()

func init() {
	// TODO: check setup, ie. is the base path writable?
	flag.StringVar(&setup.Zk, "zk", "127.0.0.1:2181", "ZK connection string")
	flag.StringVar(&setup.BasePath, "base-path", "", "Path where the locally-cached configuration will be stored")
	flag.StringVar(&setup.ZkRoot, "zk-root", "/zconfig", "ZK path to the configuration")
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
			// FIXME: here we can get a node does not exist
			fmt.Printf("main change=%v\n", change)
			err = cm.UpdateLocal()
			if err != nil {
				fmt.Printf("main change_error=%v\n", err)
			}
		case err := <-errors:
			// we'll end up with node does not exist here
			// which will kill the go routine (it's fine)
			fmt.Printf("main error=%v\n", err)
		}
	}
}

func zkConnect() *zk.Conn {
	conn, _, err := zk.Connect(strings.Split(setup.Zk, ","), setup.ZkTimeout)
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
