package main

type Setup struct {
	BasePath string
	Zk       string
	ZkRoot   string
}

func NewSetup() *Setup {
	return &Setup{}
}

func (s *Setup) Validate() bool {
	return true
}
