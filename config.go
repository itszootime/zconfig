package main

import (
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Config struct {
	data map[string]interface{}
}

func (c *Config) String() string {
	return fmt.Sprintf("%v", c.data)
}

type ConfigCache struct {
	conn     *zk.Conn
	root     string
	basePath string
}

func NewConfigCache(conn *zk.Conn, root string, basePath string) *ConfigCache {
	return &ConfigCache{conn, root, basePath}
}

func (cc *ConfigCache) Update() error {
	data, err := cc.getData(cc.root)
	if err != nil {
		return err
	}

	cfg := &Config{data: data}

	err = cc.clean(cfg)
	if err != nil {
		return err
	}

	return cc.dump(cfg)
}

func (cc *ConfigCache) clean(cfg *Config) error {
	files, err := ioutil.ReadDir(cc.basePath)
	if err != nil {
		return err
	}

	for _, file := range files {
		filename := file.Name()
		if _, ok := cfg.data[filename[:len(filename)-4]]; !ok {
			err = os.Remove(filepath.Join(cc.basePath, filename))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (cc *ConfigCache) dump(cfg *Config) error {
	for name, contents := range cfg.data {
		yml, err := yaml.Marshal(&contents)
		if err != nil {
			return err
		}
		ymlpath := filepath.Join(cc.basePath, name+".yml")
		err = ioutil.WriteFile(ymlpath, yml, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cc *ConfigCache) getData(path string) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	children, _, err := cc.conn.Children(path)
	if err != nil {
		// TODO: error variety check
		return nil, err
	}

	if len(children) > 0 {
		for i := range children {
			child, err := cc.getChildData(path, children[i])
			if err != nil {
				// TODO: error variety check
				return nil, err
			}
			data[children[i]] = child
		}
	}

	return data, nil
}

func (cc *ConfigCache) getChildData(root string, child string) (interface{}, error) {
	path := root + "/" + child
	children, _, err := cc.conn.Children(path)
	if err != nil {
		// TODO: error variety check
		return nil, err
	}

	if len(children) == 0 {
		bytes, _, err := cc.conn.Get(path)
		if err != nil {
			// TODO: error variety check
			return nil, err
		}

		if len(bytes) == 0 {
			return nil, nil
		} else {
			return string(bytes), nil
		}
	} else {
		data, err := cc.getData(path)
		if err != nil {
			// TODO: error variety check
			return nil, err
		}

		return cc.parseData(data), nil
	}
}

func (cc *ConfigCache) parseData(values map[string]interface{}) interface{} {
	valuesarr := make([]string, 0, len(values))
	isarr := true
	// if all values are nil, it's an array of values
	for k, v := range values {
		if v != nil {
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
