package watcher_test

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
	"github.com/wilhelm-murdoch/go-watcher"
)

func TestWatcher(t *testing.T) {
	var err error
	w, err := watcher.New()
	assert.Nil(t, err, "was expecting no errors, but got %s instead", err)

	err = w.AddFile("test.txt")
	assert.Nil(t, err, "was expecting no errors, but got %s instead", err)

	w.On(fsnotify.Chmod, func(file os.FileInfo, err error) error {
		fmt.Println("hi", file.Name())
		return errors.New("no.")
	})

	go func() {
		time.Sleep(4 * time.Second)
		w.Done()
	}()

	err = w.Watch()
	assert.Nil(t, err, "was expecting no errors, but got %s instead", err)
}
