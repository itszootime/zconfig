package main

import "testing"

func TestValidate(t *testing.T) {
	s := &Setup{
		Zk:       "localhost:2181",
		ZkRoot:   "/zconfig",
		BasePath: ".",
	}

	err := s.Validate()

	if err == nil {
		t.Fatalf("expected %v to be invalid, but it wasn't", s)
	}
}
