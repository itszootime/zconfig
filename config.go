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

type ConfigManager struct {
	conn     *zk.Conn
	root     string
	basePath string
}

func NewConfigManager(conn *zk.Conn, root string, basePath string) *ConfigManager {
	return &ConfigManager{conn, root, basePath}
}

func (cm *ConfigManager) UpdateLocal() error {
	data, err := cm.getData(cm.root)
	if err != nil {
		return err
	}

	cfg := &Config{data: data}

	err = cm.cleanLocal(cfg)
	if err != nil {
		return err
	}

	return cm.dumpLocal(cfg)
}

func (cm *ConfigManager) cleanLocal(cfg *Config) error {
	files, err := ioutil.ReadDir(cm.basePath)
	if err != nil {
		return err
	}

	for _, file := range files {
		filename := file.Name()
		if _, ok := cfg.data[filename[:len(filename)-4]]; !ok {
			err = os.Remove(filepath.Join(cm.basePath, filename))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (cm *ConfigManager) dumpLocal(cfg *Config) error {
	for name, contents := range cfg.data {
		yml, err := yaml.Marshal(&contents)
		if err != nil {
			return err
		}
		ymlpath := filepath.Join(cm.basePath, name+".yml")
		err = ioutil.WriteFile(ymlpath, yml, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cm *ConfigManager) getData(path string) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	children, _, err := cm.conn.Children(path)
	if err != nil {
		// TODO: what errors? maybe the error just means empty value?
		return nil, err
	}

	if len(children) > 0 {
		for i := range children {
			child, err := cm.getChildData(path, children[i])
			if err != nil {
				// TODO: errors?
				return nil, err
			}
			data[children[i]] = child
		}
	}

	return data, nil
}

func (cm *ConfigManager) getChildData(root string, child string) (interface{}, error) {
	path := root + "/" + child
	children, _, err := cm.conn.Children(path)
	if err != nil {
		// TODO: what errors? maybe the error just means empty value?
		return nil, err
	}

	if len(children) == 0 {
		// value
		bytes, _, err := cm.conn.Get(path)
		if err != nil {
			// TODO: what errors? maybe the error just means empty value?
			return nil, err
		}

		// empty values are nil
		if len(bytes) == 0 {
			return nil, nil
		} else {
			return string(bytes), nil
		}
	} else {
		// could be an array of values, or could be recursive
		data, err := cm.getData(path)
		if err != nil {
			// TODO: errors?
			return nil, err
		}

		return cm.parseData(data), nil
	}
}

// the challenge here is how to decide if this is an array
// if all values are empty strings, it's an array
// TODO: document this logic, it can be strange under certain conditions
func (cm *ConfigManager) parseData(values map[string]interface{}) interface{} {
	valuesarr := make([]string, 0, len(values))
	isarr := true
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
