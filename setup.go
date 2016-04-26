package main

import (
	"errors"
	"fmt"
	"os"
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
	return nil
}
