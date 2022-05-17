package watcher_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
	"github.com/wilhelm-murdoch/go-watcher"
)

var events = []fsnotify.Op{
	fsnotify.Write,
	fsnotify.Create,
	fsnotify.Remove,
	fsnotify.Rename,
	fsnotify.Chmod,
}

func createFile(path string) (string, error) {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return "", err
	}

	file, err := ioutil.TempFile(path, "prefix")
	if err != nil {
		return "", err
	}

	return file.Name(), nil
}

func createFiles(path string, count int) ([]string, error) {
	var files []string

	for i := 1; i <= count; i++ {
		file, err := createFile(path)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	return files, nil
}

func touchFile(path string) error {
	t := time.Now().Local()
	err := os.Chtimes(path, t, t)
	if err != nil {
		return err
	}
	return nil
}

func cleanFiles(path string) error {
	return os.RemoveAll(path)
}

func TestWatcherList(t *testing.T) {
	defer cleanFiles("./test_files")

	w, err := watcher.New()
	assert.Nil(t, err, "was expecting no errors, but got %s instead", err)

	createFiles("./test_files", 5)

	assert.Nil(t, w.AddPath("./test_files"), "was expecting no errors, but got %s instead", err)
	assert.Equal(t, 6, len(w.List()), "was expecting %d items, but got %d instead", 6, len(w.List()))
}

func TestWatcherNoCallbacks(t *testing.T) {
	w, err := watcher.New()
	assert.Nil(t, err, "was expecting no errors, but got %s instead", err)

	err = w.AddPath("watcher_test.go")
	assert.Nil(t, err, "was expecting no errors, but got %s instead", err)

	err = w.Watch()
	assert.NotNil(t, err, "was expecting no errors, but got %s instead", err)
}

func TestWatcherNoFiles(t *testing.T) {
	w, err := watcher.New()
	assert.Nil(t, err, "was expecting no errors, but got %s instead", err)

	err = w.Watch()
	assert.NotNil(t, err, "was expecting no errors, but got %s instead", err.Error())
	assert.Zero(t, len(w.List()), "was expecting zero results, but got %d instead", len(w.List()))
}

func TestWatcherWalkPath(t *testing.T) {
	defer cleanFiles("./test_files")

	w, err := watcher.New()
	assert.Nil(t, err, "was expecting no errors, but got %s instead", err)

	createFiles("./test_files", 5)
	createFiles("./test_files/sub", 5)

	err = w.WalkPath("./test_files")
	assert.Nil(t, err, "was expecting no errors, but got %s instead", err)
	assert.Equal(t, 12, len(w.List()), "was expecting %d items, but got %d instead", 12, len(w.List()))
}

func TestWatcherWalkPathInvalidFile(t *testing.T) {
	w, err := watcher.New()
	assert.Nil(t, err, "was expecting no errors, but got %s instead", err)

	err = w.WalkPath("./i-do-not-exist")
	assert.NotNil(t, err, "was expecting a pattern error, but got nothing instead")
}

func TestWatcherGlob(t *testing.T) {
	defer cleanFiles("./test_files")

	w, err := watcher.New()
	assert.Nil(t, err, "was expecting no errors, but got %s instead", err)

	createFiles("./test_files/sub", 5)

	err = w.AddGlob("./test_files/**/prefix*")
	assert.Nil(t, err, "was expecting no errors, but got %s instead", err)
	assert.Equal(t, 5, len(w.List()), "was expecting %d items, but got %d instead", 5, len(w.List()))

	err = w.Watch()
	assert.NotNil(t, err, "was expecting no errors, but got %s instead", err)
}

func TestWatcherGlobInvalidPattern(t *testing.T) {
	w, err := watcher.New()
	assert.Nil(t, err, "was expecting no errors, but got %s instead", err)

	err = w.AddGlob("a[")
	assert.NotNil(t, err, "was expecting a pattern error, but got nothing instead")
}

func TestWatcherOnCallbackSet(t *testing.T) {
	w, err := watcher.New()
	assert.Nil(t, err, "was expecting no errors, but got %s instead", err)

	callback := func(event fsnotify.Event, file os.FileInfo, err error) error {
		return nil
	}

	for _, e := range events {
		err := w.On(e, callback)
		assert.Nil(t, err, "was expecting no errors, but got %s instead", err)
	}

	const BadOp fsnotify.Op = 000
	err = w.On(BadOp, callback)
	assert.NotNil(t, err, "was expecting an error, but got nothing instead")
}

func TestWatcherAll(t *testing.T) {
	w, err := watcher.New()
	assert.Nil(t, err, "was expecting no errors, but got %s instead", err)

	w.All(func(event fsnotify.Event, file os.FileInfo, err error) error {
		return nil
	})
}

func TestWatcherWatchOn(t *testing.T) {
	defer cleanFiles("watch_path")
	var message chan string
	var finished chan bool

	w, err := watcher.New()
	assert.Nil(t, err, "was expecting no errors, but got %s instead", err)

	callback := func(event fsnotify.Event, file os.FileInfo, err error) error {
		message <- fmt.Sprintf("%s: %s", event.Name, file.Name())
		return nil
	}

	for _, e := range events {
		err := w.On(e, callback)
		assert.Nil(t, err, "was expecting no errors, but got %s instead", err)
	}

	err = os.MkdirAll("watch_path", os.ModePerm)
	assert.Nil(t, err, "was expecting no errors, but got %s instead", err)

	w.AddPath("watch_path")

	go func() {
		for {
			select {
			case m := <-message:
				fmt.Println(m)
			case <-finished:
				fmt.Println("sup")
				w.Done()
				close(message)
				close(finished)
			}
		}
	}()

	// go w.Watch()

	finished <- true
	fmt.Println("sup")
}
