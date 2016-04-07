package main

import (
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
)

type Config struct {
	data map[string]interface{}
}

func FetchConfig(conn *zk.Conn, root string) (*Config, error) {
	// TODO: fetch config data
	return &Config{}, nil
}

func (c *Config) Save(path string) error {
	// TODO: write files to path
	return nil
}

func (c *Config) String() string {
	return fmt.Sprintf("%v", c.data)
}

// // TODO: every value is a string, is this a problem? (it's hard to fix)
// func fetchValues(conn *zk.Conn, path string) (map[string]interface{}, error) {
//   v := make(map[string]interface{})

//   // get children
//   children, _, err := conn.Children(path)
//   if err != nil {
//     // TODO: what errors? maybe the error just means empty value?
//     return nil, err
//   }

//   if len(children) > 0 {
//     for i := range children {
//       childpath := path + "/" + children[i]
//       childchildren, _, err := conn.Children(childpath)
//       if err != nil {
//         // TODO: what errors? maybe the error just means empty value?
//         return nil, err
//       }

//       if len(childchildren) == 0 {
//         // value
//         bytes, _, err := conn.Get(childpath)
//         if err != nil {
//           // TODO: what errors? maybe the error just means empty value?
//           return nil, err
//         }
//         v[children[i]] = string(bytes)
//       } else {
//         // could be an array of values, or could be recursive
//         childvalues, err := fetchValues(conn, childpath)
//         if err != nil {
//           // TODO: errors
//           return nil, err
//         }

//         // the challenge here is how to decide if this is an array
//         // if all values are empty strings, it's an array
//         // TODO: document this logic, it can be strange under certain conditions
//         valuesarr := make([]string, 0, len(childvalues))
//         isarr := true
//         for k, v := range childvalues {
//           // TODO: seems hacky
//           if len(fmt.Sprintf("%v", v)) > 0 {
//             isarr = false
//             break
//           }
//           valuesarr = append(valuesarr, k)
//         }

//         if isarr {
//           v[children[i]] = valuesarr
//         } else {
//           v[children[i]] = childvalues
//         }
//       }
//     }
//   }

//   return v, nil
// }
