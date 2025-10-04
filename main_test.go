package main

import (
	"os"
	"testing"
)

func TestRun(t *testing.T) {
	os.Args = []string{"cmd", "items.xml"}
	err := run()
	if err != nil {
		t.Fatalf("Failed to run: %v", err)
	}
	t.Logf("Run succeeded")
}
