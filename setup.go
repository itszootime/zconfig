package main

import (
	"time"
)

type Setup struct {
	BasePath string
	Zk       string
	ZkRoot   string
	ZkTimeout time.Duration
}

func NewSetup() *Setup {
	return &Setup{ZkTimeout: time.Second}
}

func (s *Setup) Validate() bool {
	return true
}
