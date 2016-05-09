package main

import (
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

type mockConn struct {
	children map[string][]string
	values   map[string][]byte
}

func (m *mockConn) SetGet(path string, value []byte) {
	if m.values == nil {
		m.values = make(map[string][]byte)
	}
	m.values[path] = value
}

func (m *mockConn) SetChildren(path string, children []string) {
	if m.children == nil {
		m.children = make(map[string][]string)
	}
	m.children[path] = children
}

func (m *mockConn) Children(path string) ([]string, *zk.Stat, error) {
	return m.children[path], nil, nil
}

func (m *mockConn) Get(path string) ([]byte, *zk.Stat, error) {
	return m.values[path], nil, nil
}

func assertCacheEquals(t *testing.T, cc *ConfigCache, exp map[string]interface{}) {
	for name, value := range exp {
		path := filepath.Join(cc.basePath, name+".yml")
		actb, err := ioutil.ReadFile(path)
		if err != nil {
			t.Fatalf("expected file %v couldn't be read", path)
		}
		act := string(actb)

		expb, err := yaml.Marshal(&value)
		if err != nil {
			t.Fatalf("expected data couldn't be marshalled")
		}
		exp := string(expb)

		if act != exp {
			t.Fatalf("got: %v, expected: %v", act, exp)
		}
	}
}

var root = "/zconfig"
var base = filepath.Join(os.TempDir(), "zconfig")

func TestMain(m *testing.M) {
	err := os.Mkdir(base, 0777)
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't create %v: %v\n", base, err)
		os.Exit(1)
	}
	code := m.Run()
	err = os.RemoveAll(base)
	if err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "couldn't remove %v: %v\n", base, err)
	}
	os.Exit(code)
}

func TestUpdate(t *testing.T) {
	conn := mockConn{}
	cache := NewConfigCache(&conn, root, base)

	conn.SetChildren(root, []string{"servers"})
	conn.SetGet(root+"/servers", []byte(""))
	conn.SetChildren(root+"/servers", []string{})

	cache.Update()

	assertCacheEquals(t, cache, map[string]interface{}{
		"servers": nil,
	})

	conn.SetChildren(root+"/servers", []string{"db"})
	conn.SetGet(root+"/servers/db", []byte(""))
	conn.SetChildren(root+"/servers/db", []string{})

	cache.Update()

	assertCacheEquals(t, cache, map[string]interface{}{
		"servers": []string{"db"},
	})

	conn.SetChildren(root+"/servers", []string{"db", "timeout"})
	conn.SetGet(root+"/servers/timeout", []byte("1000"))
	conn.SetChildren(root+"/servers/timeout", []string{})

	cache.Update()

	assertCacheEquals(t, cache, map[string]interface{}{
		"servers": map[string]interface{}{
			"db":      nil,
			"timeout": "1000",
		},
	})

	conn.SetChildren(root+"/servers/db", []string{"192.168.0.1", "192.168.0.2"})
	conn.SetGet(root+"/servers/db/192.168.0.1", []byte(""))
	conn.SetChildren(root+"/servers/db/192.168.0.1", []string{})
	conn.SetGet(root+"/servers/db/192.168.0.2", []byte(""))
	conn.SetChildren(root+"/servers/db/192.168.0.2", []string{})

	cache.Update()

	assertCacheEquals(t, cache, map[string]interface{}{
		"servers": map[string]interface{}{
			"db":      []string{"192.168.0.1", "192.168.0.2"},
			"timeout": "1000",
		},
	})
}

func TestUpdateEmpty(t *testing.T) {
	conn := mockConn{}
	conn.SetChildren(root, []string{})

	cache := NewConfigCache(&conn, root, base)
	cache.Update()

	files, _ := ioutil.ReadDir(base)
	if len(files) > 0 {
		t.Fatalf("expected %v to be empty, but it wasn't", base)
	}
}
