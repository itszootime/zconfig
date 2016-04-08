package main

import (
	"flag"
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"strings"
	"time"
)

type Setup struct {
	BasePath string
	Zk       string
	ZkRoot   string
}

var setup = Setup{}

func init() {
	flag.StringVar(&setup.BasePath, "base-path", "", "base path")
	flag.StringVar(&setup.Zk, "zk", "127.0.0.1:2181", "zk")
	flag.StringVar(&setup.ZkRoot, "zk-root", "/zconfig", "zk root")
}

func iferr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()

	conn := zkConnect()
	defer conn.Close()
	zkInit(conn)

	printValues(conn, setup.ZkRoot)

	// get and watch root
	watcher := NewWatcher(conn, setup.ZkRoot)
	changes, errors := watcher.Start()
	for {
		select {
		case change := <-changes:
			fmt.Printf("main:change path=%v\n", change)
			printValues(conn, setup.ZkRoot)
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

func printValues(conn *zk.Conn, path string) {
	config, err := FetchConfig(conn, path)
	iferr(err)
	fmt.Printf("%v\n", config)
}
