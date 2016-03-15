package main

import (
  "fmt"
  "github.com/samuel/go-zookeeper/zk"
  "strings"
  "time"
)

// struct Setup {
//   zkRoot
//   basePath
// }

func iferr(err error) {
  if err != nil {
    panic(err)
  }
}

func connect() *zk.Conn {
  str := "localhost:2181"
  conn, _, err := zk.Connect(strings.Split(str, ","), time.Second)
  iferr(err)
  return conn
}

func mirror(conn *zk.Conn, path string) (chan []string, chan error) {
  snapshots := make(chan []string)
  errors := make(chan error)

  // we have two different types of watches: ChildrenW and GetW
  go func() {
    for {
      // ChildrenW is only for immediate children of root
      // we would then need to watch each of these children...
      snapshot, _, events, err := conn.ChildrenW(path)
      if err != nil {
        errors <- err
        return
      }
      snapshots <- snapshot
      evt := <-events
      if evt.Err != nil {
        errors <- evt.Err
        return
      }
    }
  }()
  return snapshots, errors
}

func main() {
  conn := connect()
  defer conn.Close()

  flags := int32(0)
  acl := zk.WorldACL(zk.PermAll)  

  // make root if it doesn't exist
  zkRoot := "/zconfig"
  // will get node already exists without this
  // BUT i know this isn't correct zk usage - what happens if someone else
  // creates the node inbetween the exists and create calls?
  exists, _, err := conn.Exists(zkRoot)
  iferr(err)
  if !exists {
    _, err := conn.Create(zkRoot, []byte("helloroot"), flags, acl)
    iferr(err)
  }

  // get and watch root
  snapshots, errors := mirror(conn, zkRoot)
  for {
    select {
    case snapshot := <-snapshots:
      fmt.Printf("%+v\n", snapshot)
    case err := <-errors:
      panic(err)
    }
  }
}
