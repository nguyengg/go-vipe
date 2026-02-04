package main

import (
	"io"
	"log"
	"os"

	"github.com/nguyengg/go-vipe"
)

func main() {
	log.SetFlags(0)

	f, err := vipe.From(os.Stdin, os.Args[1:]...)
	if err != nil {
		log.Fatal(err)
	}
	defer func(f *os.File) {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}(f)

	_, err = io.Copy(os.Stdout, f)
	_ = f.Close()
	_ = os.Remove(f.Name())
	if err != nil {
		log.Fatal(err)
	}
}
