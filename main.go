package main

import (
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/samuel/go-zookeeper/zk"
	"os"
	"strings"
)

var setup = NewSetup()

type zkLogger struct {
}

func (l *zkLogger) Printf(msg string, args ...interface{}) {
	log.Debug(fmt.Sprintf(msg, args...))
}

func init() {
	log.SetLevel(log.InfoLevel)

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

func main() {
	conn := zkConnect()
	defer conn.Close()
	zkInit(conn)

	cm := NewConfigManager(conn, setup.ZkRoot, setup.BasePath)
	err := cm.UpdateLocal()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Couldn't update local configuration cache")
	}

	watcher := NewWatcher(conn, setup.ZkRoot)
	changes, errors := watcher.Start()

	for {
		select {
		case change := <-changes:
			log.WithFields(log.Fields{
				"path": change,
			}).Info("Change observed")

			err = cm.UpdateLocal()
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("Couldn't update local configuration cache")
			}
		case err := <-errors:
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Got error from watcher")
		}
	}
}

func zkConnect() *zk.Conn {
	conn, _, err := zk.Connect(strings.Split(setup.Zk, ","), setup.ZkTimeout)
	conn.SetLogger(&zkLogger{})
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Couldn't create ZK connection")
	}
	return conn
}

func zkInit(conn *zk.Conn) {
	exists, _, err := conn.Exists(setup.ZkRoot)
	if err != nil {
		log.WithFields(log.Fields{
			"root":  setup.ZkRoot,
			"error": err,
		}).Fatal("Couldn't initialise ZK")
	}

	if !exists {
		acl := zk.WorldACL(zk.PermAll)
		_, err := conn.Create(setup.ZkRoot, nil, int32(0), acl)
		if err != nil && err != zk.ErrNodeExists {
			log.WithFields(log.Fields{
				"root":  setup.ZkRoot,
				"error": err,
			}).Fatal("Couldn't create ZK root")
		}
	}
}
