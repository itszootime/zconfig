package main

/*
 * First version:
 *
 * - a method to generate configs from roots
 * - a shared channel for node updates, just publishes the root path (i.e.
 *   /zconfig/servers)
 * - watch children on super-root
 * - watch children and values on non-root (and so on)
 * - keep track of no. watches
 */

import (
  "flag"
  "fmt"
  "github.com/samuel/go-zookeeper/zk"
  "strings"
  "time"
  "encoding/json"
)


type Setup struct {
  basePath string
  zk string
  zkRoot string
}

var setup = Setup{}

func init() {
  flag.StringVar(&setup.basePath, "base-path", "", "base path")
  flag.StringVar(&setup.zk, "zk", "127.0.0.1:2181", "zk")
  flag.StringVar(&setup.zkRoot, "zk-root", "/zconfig", "zk root")
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

  printValues(conn, setup.zkRoot)

  // get and watch root
  changes, errors := watch(conn, setup.zkRoot)
  for {
    select {
    case change := <-changes:
      fmt.Printf("main:change path=%v\n", change)
      printValues(conn, setup.zkRoot)
    case <-errors:
      // we'll end up with node does not exist here
      // which will kill the go routine (it's fine)
      // fmt.Printf("main:error error=%v\n", err)
    }
  }
}

func zkConnect() *zk.Conn {
  // TODO: allow configurable timeout?
  conn, _, err := zk.Connect(strings.Split(setup.zk, ","), time.Second)
  iferr(err) // severe
  return conn
}

func zkInit(conn *zk.Conn) {
  // TODO: ensure these flags are correct
  flags := int32(0)
  acl := zk.WorldACL(zk.PermAll)

  exists, _, err := conn.Exists(setup.zkRoot)
  iferr(err) // severe
  if !exists {
    // TODO: ignore node already exists here
    _, err := conn.Create(setup.zkRoot, nil, flags, acl)
    iferr(err)
  }
}

// TODO: move!
// // cache caches on disk the configuration values stored at the given root.
// func cache(conn *zk.Conn, root string) error {
//   values, err := fetchValues(conn, root)
//   if (err) {
//     return err
//   }

//   // now write to YAML
//   // needs a base path
// }

// TODO: every value is a string, is this a problem? (it's hard to fix)
func fetchValues(conn *zk.Conn, path string) (map[string]interface{}, error) {
  v := make(map[string]interface{})

  // get children
  children, _, err := conn.Children(path)
  if err != nil {
    // TODO: what errors? maybe the error just means empty value?
    return nil, err
  }

  if len(children) > 0 {
    for i := range children {
      childpath := path + "/" + children[i]
      childchildren, _, err := conn.Children(childpath)
      if err != nil {
        // TODO: what errors? maybe the error just means empty value?
        return nil, err
      }

      if len(childchildren) == 0 {
        // value
        bytes, _, err := conn.Get(childpath)
        if err != nil {
          // TODO: what errors? maybe the error just means empty value?
          return nil, err
        }
        v[children[i]] = string(bytes)
      } else {
        // could be an array of values, or could be recursive
        childvalues, err := fetchValues(conn, childpath)
        if err != nil {
          // TODO: errors
          return nil, err
        }

        // the challenge here is how to decide if this is an array
        // if all values are empty strings, it's an array
        // TODO: document this logic, it can be strange under certain conditions
        valuesarr := make([]string, 0, len(childvalues))
        isarr := true
        for k, v := range childvalues {
          // TODO: seems hacky
          if len(fmt.Sprintf("%v", v)) > 0 {
            isarr = false
            break
          }
          valuesarr = append(valuesarr, k)
        }

        if isarr {
          v[children[i]] = valuesarr
        } else {
          v[children[i]] = childvalues
        }
      }
    }
  }

  return v, nil
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

func printValues(conn *zk.Conn, path string) {
  values, err := fetchValues(conn, path)
  iferr(err)
  bytes, err := json.Marshal(values)
  iferr(err)
  fmt.Printf("%v\n", string(bytes))
}

func watch(conn *zk.Conn, path string) (chan string, chan error) {
  changes := make(chan string)
  errors := make(chan error)

  go func() {
    watchTree(conn, path, changes, errors)
  }()

  return changes, errors
}
