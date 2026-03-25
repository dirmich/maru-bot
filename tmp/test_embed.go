package main

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"
)

//go:embed dist
var testAssets embed.FS

func main() {
	distFS, err := fs.Sub(testAssets, "dist")
	if err != nil {
		fmt.Printf("Sub error: %v\n", err)
		return
	}

	paths := []string{
		"index.html",
		"assets/index-hz622eoM.js",
		"assets/index-DOpi7LuF.css",
		"/index.html",
		"/assets/index-hz622eoM.js",
	}

	for _, p := range paths {
		target := strings.TrimPrefix(p, "/")
		f, err := distFS.Open(target)
		if err != nil {
			fmt.Printf("Open(%q) [target: %q] -> FAIL: %v\n", p, target, err)
		} else {
			stat, _ := f.Stat()
			fmt.Printf("Open(%q) [target: %q] -> SUCCESS (size: %d, isDir: %v)\n", p, target, stat.Size(), stat.IsDir())
			f.Close()
		}
	}
}
