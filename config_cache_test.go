package main

import (
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"github.com/stretchr/testify/mock"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

type mockConn struct {
	mock.Mock
}

func (m *mockConn) Children(path string) ([]string, *zk.Stat, error) {
	args := m.Called(path)
	return args.Get(0).([]string), nil, nil
}

func (m *mockConn) Get(path string) ([]byte, *zk.Stat, error) {
	args := m.Called(path)
	return args.Get(0).([]byte), nil, nil
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
	conn.On("Children", root).Return([]string{"servers"})
	conn.On("Get", root+"/servers").Return([]byte(""))
	conn.On("Children", root+"/servers").Return([]string{})

	cache := NewConfigCache(&conn, root, base)
	cache.Update()

	assertCacheEquals(t, cache, map[string]interface{}{
		"servers": nil,
	})

	conn = mockConn{}
	conn.On("Children", root).Return([]string{"servers"})
	conn.On("Get", root+"/servers").Return([]byte(""))
	conn.On("Children", root+"/servers").Return([]string{"db"})
	conn.On("Get", root+"/servers/db").Return([]byte(""))
	conn.On("Children", root+"/servers/db").Return([]string{})

	cache = NewConfigCache(&conn, root, base)
	cache.Update()

	assertCacheEquals(t, cache, map[string]interface{}{
		"servers": []string{"db"},
	})

	conn = mockConn{}
	conn.On("Children", root).Return([]string{"servers"})
	conn.On("Get", root+"/servers").Return([]byte(""))
	conn.On("Children", root+"/servers").Return([]string{"db", "timeout"})
	conn.On("Get", root+"/servers/db").Return([]byte(""))
	conn.On("Children", root+"/servers/db").Return([]string{})
	conn.On("Get", root+"/servers/timeout").Return([]byte("1000"))
	conn.On("Children", root+"/servers/timeout").Return([]string{})

	cache = NewConfigCache(&conn, root, base)
	cache.Update()

	assertCacheEquals(t, cache, map[string]interface{}{
		"servers": map[string]interface{}{
			"db":      nil,
			"timeout": "1000",
		},
	})

	conn = mockConn{}
	conn.On("Children", root).Return([]string{"servers"})
	conn.On("Get", root+"/servers").Return([]byte(""))
	conn.On("Children", root+"/servers").Return([]string{"db", "timeout"})
	conn.On("Get", root+"/servers/db").Return([]byte(""))
	conn.On("Children", root+"/servers/db").Return([]string{"192.168.0.1", "192.168.0.2"})
	conn.On("Get", root+"/servers/db/192.168.0.1").Return([]byte(""))
	conn.On("Children", root+"/servers/db/192.168.0.1").Return([]string{})
	conn.On("Get", root+"/servers/db/192.168.0.2").Return([]byte(""))
	conn.On("Children", root+"/servers/db/192.168.0.2").Return([]string{})
	conn.On("Get", root+"/servers/timeout").Return([]byte("1000"))
	conn.On("Children", root+"/servers/timeout").Return([]string{})

	cache = NewConfigCache(&conn, root, base)
	cache.Update()

	assertCacheEquals(t, cache, map[string]interface{}{
		"servers": map[string]interface{}{
			"db":      []string{"192.168.0.1", "192.168.0.2"},
			"timeout": "1000",
		},
	})

	conn = mockConn{}
	conn.On("Children", root).Return([]string{})

	cache = NewConfigCache(&conn, root, base)
	cache.Update()

	files, _ := ioutil.ReadDir(base)
	if len(files) > 0 {
		t.Fatalf("expected %v to be empty, but it wasn't", base)
	}
}
