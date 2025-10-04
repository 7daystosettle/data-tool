package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/7daystosettle/data-tool/ko"
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
	start := time.Now()
	if len(os.Args) < 3 {
		fmt.Printf("usage: %s <src_path> <out_path>\n", os.Args[0])
		os.Exit(1)
	}
	inPath := os.Args[1]
	outPath := os.Args[2]

	info, err := os.Stat(inPath)
	if err != nil {
		return fmt.Errorf("stat input path: %w", err)
	}

	if !info.IsDir() {
		err := convert(inPath, outPath)
		if err != nil {
			return fmt.Errorf("convert: %w", err)
		}
		return nil
	}

	totalConverted := 0

	files, err := os.ReadDir(inPath)
	if err != nil {
		return fmt.Errorf("read input dir: %w", err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		ext := filepath.Ext(file.Name())
		if ext != ".xml" && ext != ".kdl" {
			continue
		}
		inFile := filepath.Join(inPath, file.Name())
		var outFile string
		if ext == ".xml" {
			outFile = filepath.Join(outPath, file.Name()[:len(file.Name())-len(ext)]+".kdl")
		} else {
			outFile = filepath.Join(outPath, file.Name()[:len(file.Name())-len(ext)]+".xml")
		}
		err := convert(inFile, outFile)
		if err != nil {
			fmt.Printf("Failed to convert %s: %v\n", inFile, err)
		}

		totalConverted++
	}

	fmt.Printf("Converted %d files in %0.2f seconds\n", totalConverted, time.Since(start).Seconds())

	return nil
}

func convert(inPath, outPath string) error {

	r, err := os.Open(inPath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer r.Close()

	var doc *ko.Ko

	inExt := filepath.Ext(inPath)
	switch inExt {
	case ".xml":
		doc, err = ko.NewFromXml(r)
		if err != nil {
			return fmt.Errorf("converting xml to kdl: %w", err)
		}

	case ".kdl":
		doc, err = ko.NewFromKdl(r)
		if err != nil {
			return fmt.Errorf("converting kdl to xml: %w", err)
		}
	default:
		return fmt.Errorf("unsupported input file extension: %s", inExt)
	}

	w, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create %s: %w", filepath.Base(outPath), err)
	}
	defer w.Close()

	switch filepath.Ext(outPath) {
	case ".xml":
		err = doc.ToXml(w)
		if err != nil {
			return fmt.Errorf("writing xml file: %w", err)
		}
	case ".kdl":
		err = doc.ToKdl(w)
		if err != nil {
			return fmt.Errorf("writing kdl file: %w", err)
		}
	default:
		return fmt.Errorf("unsupported output file extension: %s", filepath.Ext(outPath))
	}

	return nil
}
