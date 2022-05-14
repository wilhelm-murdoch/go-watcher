package watcher_test

import (
	"fmt"
	"testing"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
	"github.com/wilhelm-murdoch/go-watcher"
)

func TestWatcher(t *testing.T) {
	w, err := watcher.New()
	assert.Nil(t, err, "was expecting no errors, but got %s instead", err)

	err = w.AddFile("test.txt")
	assert.Nil(t, err, "was expecting no errors, but got %s instead", err)

	w.On(fsnotify.Write, func(file string) error {
		fmt.Println("hi")
		w.Done()
		return nil
	})

	err = w.Watch()
	assert.Nil(t, err, "was expecting no errors, but got %s instead", err)
}
