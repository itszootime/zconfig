package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type Setup struct {
	BasePath  string
	Zk        string
	ZkRoot    string
	ZkTimeout time.Duration
}

func NewSetup() *Setup {
	return &Setup{ZkTimeout: time.Second}
}

func (s *Setup) Validate() error {
	if _, err := os.Stat(s.BasePath); err != nil {
		return errors.New(fmt.Sprintf("Invalid base path! %v", err.Error()))
	}

	files, err := ioutil.ReadDir(s.BasePath)
	if err != nil {
		return errors.New(fmt.Sprintf("Invalid base path! %v", err.Error()))
	}
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".yml") {
			return errors.New(fmt.Sprintf("Non-ZConfig file %v detected in base path, please remove file or choose another directory", f.Name()))
		}
	}

	return nil
}
