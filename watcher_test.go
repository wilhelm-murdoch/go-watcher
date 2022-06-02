package watcher_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
	"github.com/wilhelm-murdoch/go-watcher"
)

const (
	tmpDir = "watch_path"
)

var (
	events = []fsnotify.Op{
		fsnotify.Write,
		fsnotify.Create,
		fsnotify.Remove,
		fsnotify.Rename,
		fsnotify.Chmod,
	}

	tests = []struct {
		Name    string
		Event   fsnotify.Op
		Trigger func()
	}{
		{
			Name:  "TestOnWrite",
			Event: fsnotify.Write,
			Trigger: func() {
				appendToFile(filepath.Join(tmpDir, "test_exists.txt"))
			},
		},
		{
			Name:  "TestOnRename",
			Event: fsnotify.Rename,
			Trigger: func() {
				os.Rename(filepath.Join(tmpDir, "test_rename.txt"), filepath.Join(tmpDir, "test_renamed.txt"))
			},
		},
		{
			Name:  "TestOnCreate",
			Event: fsnotify.Create,
			Trigger: func() {
				appendToFile(filepath.Join(tmpDir, "test_created.txt"))
			},
		},
		{
			Name:  "TestOnChmod",
			Event: fsnotify.Chmod,
			Trigger: func() {
				touchFile(filepath.Join(tmpDir, "test_exists.txt"))
			},
		},
		{
			Name:  "TestOnRemove",
			Event: fsnotify.Remove,
			Trigger: func() {
				os.Remove(filepath.Join(tmpDir, "test_delete.txt"))
			},
		},
	}
)

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

func appendToFile(path string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	f.Sync()
	defer f.Close()

	if _, err := f.WriteString("hello world\n"); err != nil {
		return err
	}

	return nil
}

func TestWatcherList(t *testing.T) {
	defer cleanFiles(tmpDir)

	w, err := watcher.New()
	assert.Nil(t, err, "was expecting no errors, but got %s instead", err)

	createFiles(tmpDir, 5)

	assert.Nil(t, w.AddPath(tmpDir), "was expecting no errors, but got %s instead", err)
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
	defer cleanFiles(tmpDir)

	w, err := watcher.New()
	assert.Nil(t, err, "was expecting no errors, but got %s instead", err)

	createFiles(tmpDir, 5)
	createFiles(filepath.Join(tmpDir, "sub"), 5)

	err = w.WalkPath(tmpDir)
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
	defer cleanFiles(tmpDir)

	w, err := watcher.New()
	assert.Nil(t, err, "was expecting no errors, but got %s instead", err)

	createFiles(filepath.Join(tmpDir, "sub"), 5)

	err = w.AddGlob(filepath.Join(tmpDir, "**/prefix*"))
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

func TestWatcherWatchReturnError(t *testing.T) {
	teardownTests := setupTests(t)
	defer teardownTests(t)

	w, err := watcher.New()
	assert.Nil(t, err, "was expecting no errors, but got %s instead", err)

	w.AddPath(tmpDir)

	err = w.On(fsnotify.Rename, func(event fsnotify.Event, file os.FileInfo, err error) error {
		return errors.New("welp")
	})
	assert.Nil(t, err, "was expecting no errors, but got %s instead", err)

	os.Rename(filepath.Join(tmpDir, "test_rename.txt"), filepath.Join(tmpDir, "test_renamed.txt"))

	err = w.Watch()
	assert.NotNil(t, err, "was expecting an error, but got nothing instead")
}

func setupTests(t *testing.T) func(*testing.T) {
	err := os.RemoveAll(tmpDir)
	assert.Nil(t, err, "was expecting no errors, but got %s instead", err)

	err = os.MkdirAll(tmpDir, os.ModePerm)
	assert.Nil(t, err, "was expecting no errors, but got %s instead", err)

	files := []string{
		"test_exists.txt",
		"test_rename.txt",
		"test_delete.txt",
	}

	for _, file := range files {
		f, err := os.OpenFile(filepath.Join(tmpDir, file), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		assert.Nil(t, err, "was expecting no errors, but got %s instead", err)
		f.Close()
	}

	return func(t *testing.T) {
		err := os.RemoveAll(tmpDir)
		assert.Nil(t, err, "was expecting no errors, but got %s instead", err)
	}
}

func TestWatcherWatchAllSuite(t *testing.T) {
	teardownTests := setupTests(t)
	defer teardownTests(t)

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			w, err := watcher.New()
			assert.Nil(t, err, "was expecting no errors, but got %s instead", err)

			w.AddPath(tmpDir)

			w.All(func(event fsnotify.Event, file os.FileInfo, err error) error {
				if event.Op&test.Event == test.Event {
					w.Done()
				}
				return nil
			})
			assert.Nil(t, err, "was expecting no errors, but got %s instead", err)

			test.Trigger()

			w.Watch()
		})
	}
}

func TestWatcherWatchOnSuite(t *testing.T) {
	teardownTests := setupTests(t)
	defer teardownTests(t)

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			w, err := watcher.New()
			assert.Nil(t, err, "was expecting no errors, but got %s instead", err)

			w.AddPath(tmpDir)

			err = w.On(test.Event, func(event fsnotify.Event, file os.FileInfo, err error) error {
				if event.Op&test.Event == test.Event {
					w.Done()
				}
				return nil
			})
			assert.Nil(t, err, "was expecting no errors, but got %s instead", err)

			test.Trigger()

			w.Watch()
		})
	}
}
