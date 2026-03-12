package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run scripts/zip_pack.go <destination.zip> <source_dir_or_file1> [source2...]")
		os.Exit(1)
	}

	zipPath := os.Args[1]
	sources := os.Args[2:]

	if err := createZip(zipPath, sources); err != nil {
		fmt.Printf("Error creating zip: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Zip created: %s\n", zipPath)
}

func createZip(zipPath string, sources []string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	for _, source := range sources {
		info, err := os.Stat(source)
		if err != nil {
			return err
		}

		var baseDir string
		if info.IsDir() {
			baseDir = filepath.Dir(source)
		} else {
			baseDir = filepath.Dir(source)
		}

		err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}

			// relPath for inside the zip
			relPath := strings.TrimPrefix(path, baseDir)
			relPath = strings.TrimPrefix(relPath, string(filepath.Separator))
			
			// Always use forward slashes in zip
			header.Name = filepath.ToSlash(relPath)

			if info.IsDir() {
				header.Name += "/"
			} else {
				header.Method = zip.Deflate
			}

			writer, err := archive.CreateHeader(header)
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(writer, file)
			return err
		})
		if err != nil {
			return err
		}
	}

	return nil
}
