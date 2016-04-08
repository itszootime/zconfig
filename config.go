package main

import (
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
)

type Config struct {
	data map[string]interface{}
}

func FetchConfig(conn *zk.Conn, root string) (*Config, error) {
	data, err := getData(conn, root)
	return &Config{data: data}, err
}

func (c *Config) Save(path string) error {
	// TODO: write files to path
	return nil
}

func (c *Config) String() string {
	return fmt.Sprintf("%v", c.data)
}

func getData(conn *zk.Conn, path string) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	children, _, err := conn.Children(path)
	if err != nil {
		// TODO: what errors? maybe the error just means empty value?
		return nil, err
	}

	if len(children) > 0 {
		for i := range children {
			child, err := getChildData(conn, path, children[i])
			if err != nil {
				// TODO: errors?
				return nil, err
			}
			data[children[i]] = child
		}
	}

	return data, nil
}

func getChildData(conn *zk.Conn, root string, child string) (interface{}, error) {
	path := root + "/" + child
	children, _, err := conn.Children(path)
	if err != nil {
		// TODO: what errors? maybe the error just means empty value?
		return nil, err
	}

	if len(children) == 0 {
		// value
		bytes, _, err := conn.Get(path)
		if err != nil {
			// TODO: what errors? maybe the error just means empty value?
			return nil, err
		}
		return string(bytes), nil
	} else {
		// could be an array of values, or could be recursive
		data, err := getData(conn, path)
		if err != nil {
			// TODO: errors?
			return nil, err
		}

		return parseData(data), nil
	}
}

// the challenge here is how to decide if this is an array
// if all values are empty strings, it's an array
// TODO: document this logic, it can be strange under certain conditions
func parseData(values map[string]interface{}) interface{} {
	valuesarr := make([]string, 0, len(values))
	isarr := true
	for k, v := range values {
		// TODO: seems hacky
		if len(fmt.Sprintf("%v", v)) > 0 {
			isarr = false
			break
		}
		valuesarr = append(valuesarr, k)
	}

	if isarr {
		return valuesarr
	} else {
		return values
	}
}
