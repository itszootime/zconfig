package main

import (
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
)

type Config struct {
	data map[string]interface{}
}

func (c *Config) Save(path string) error {
	for name, contents := range c.data {
		yml, err := yaml.Marshal(&contents)
		if err != nil {
			return err
		}
		ymlpath := filepath.Join(path, name+".yml")
		err = ioutil.WriteFile(ymlpath, yml, 0644)
		if err != nil {
			return err
		}
	}
	return nil
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

func (cm *ConfigManager) GetConfig() (*Config, error) {
	data, err := cm.getData(cm.root)
	return &Config{data: data}, err
}

func (cm *ConfigManager) UpdateLocal() error {
	config, err := cm.GetConfig()
	if err != nil {
		return err
	}
	return config.Save(cm.basePath)
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
		return string(bytes), nil
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
		if str, ok := v.(string); !ok || len(str) > 0 {
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
