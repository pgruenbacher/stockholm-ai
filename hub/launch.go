package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func link(o, n string) {
	if err := os.Symlink(o, n); err != nil && !strings.Contains(err.Error(), "file exists") {
		panic(err)
	}
}

func main() {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic(fmt.Errorf("Could not find the running file"))
	}
	dir, err := filepath.Abs(filepath.Dir(file))
	if err != nil {
		panic(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "deps", "src"), os.FileMode(0755)); err != nil && err.Error() != "file exists" {
		panic(err)
	}
	link(filepath.Join(filepath.Dir(dir), "github.com"), filepath.Join(dir, "deps", "src", "github.com"))
	link(filepath.Join(filepath.Dir(dir), "code.google.com"), filepath.Join(dir, "deps", "src", "code.google.com"))
	link(filepath.Dir(dir), filepath.Join(filepath.Dir(dir), "github.com", "zond", "stockholm-ai"))
}