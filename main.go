package main

import (
	"flag"
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"os"
	"strings"
)

var setup = NewSetup()

func init() {
	flag.StringVar(&setup.Zk, "zk", "127.0.0.1:2181", "ZK connection string")
	flag.StringVar(&setup.BasePath, "base-path", ".", "Path where the locally-cached configuration will be stored")
	flag.StringVar(&setup.ZkRoot, "zk-root", "/zconfig", "ZK path to the configuration")
	flag.Parse()

	if err := setup.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n\n", err.Error())
		flag.Usage()
		os.Exit(1)
	}
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

	watcher := NewWatcher(conn, setup.ZkRoot)
	changes, errors := watcher.Start()

	for {
		select {
		case change := <-changes:
			fmt.Printf("main change=%v\n", change)
			err = cm.UpdateLocal()
			if err != nil {
				fmt.Printf("main change_error=%v\n", err)
			}
		case err := <-errors:
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

	_, err := conn.Create(setup.ZkRoot, nil, flags, acl)
	if err != nil && err != zk.ErrNodeExists {
		panic(err)
	}
}
