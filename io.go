package main

import (
	"log"
	"os"
	"path/filepath"
)

type Closeable interface {
	Close()
}

type Scope struct {
	f *os.File
}

func (t Scope) Close() {
	t.f.Close()
	log.SetOutput(os.Stdout)
}

func to_disk(dir string, file string) Scope {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			if os.MkdirAll(dir, 0777) != nil {
				log.Fatal("error making a dir")
			}
		}
	}
	file = filepath.Join(dir, file)
	if _, err := os.Stat(dir + "/" + file); err == nil {
		os.Remove(file)
	}
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	log.SetOutput(f)
	return Scope{f}
}
