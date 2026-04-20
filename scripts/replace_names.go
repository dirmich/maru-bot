package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	root := "../maru-bot"
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == "releases" || info.Name() == "dist" {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip binary files by extension
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".exe" || ext == ".png" || ext == ".ico" || ext == ".jpg" || ext == ".dmg" {
			return nil
		}

		content, err := ioutil.ReadFile(path)
		if err != nil {
			return nil
		}

		newContent := bytes.ReplaceAll(content, []byte("maruminibot"), []byte("marubot"))
		newContent = bytes.ReplaceAll(newContent, []byte("MaruMiniBot"), []byte("MaruBot"))

		if !bytes.Equal(content, newContent) {
			fmt.Printf("Updating: %s\n", path)
			err = ioutil.WriteFile(path, newContent, info.Mode())
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("Replacement complete!")
	}
}
