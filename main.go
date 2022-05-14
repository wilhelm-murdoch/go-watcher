package main

import (
	"fmt"
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/wilhelm-murdoch/go-watcher/watcher"
)

func main() {
	w, _ := watcher.New()
	w.AddGlob("./**/*.go")
	w.AddGlob("./*.go")

	w.All(func(event fsnotify.Event, file string) error {
		fmt.Println("something happened to:", file)
		return nil
	})

	if err := w.Watch(); err != nil {
		log.Fatal(err)
	}
}
