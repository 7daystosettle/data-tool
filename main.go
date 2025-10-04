package main

import (
	"fmt"
	"os"

	"github.com/7daystosettle/data-tool/dt"
)

func main() {
	err := run()
	if err != nil {
		fmt.Printf("Failed to run: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("missing path argument")
	}
	path := os.Args[1]
	fmt.Printf("Path: %s\n", path)

	info, err := dt.FromXML(path)
	if err != nil {
		return fmt.Errorf("from xml: %w", err)
	}
	os.WriteFile("info.txt", []byte(info.String()), 0644)

	return nil
}
