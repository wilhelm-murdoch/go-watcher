package watcher

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"golang.org/x/sync/errgroup"
)

type Watcher struct {
	fsnotify                                       *fsnotify.Watcher
	done                                           chan bool
	onRemove, onCreate, onWrite, onRename, onChmod func(string) error
	onAll                                          func(fsnotify.Event, string) error
}

func New() (*Watcher, error) {
	fsn, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Watcher{
		fsnotify: fsn,
		done:     make(chan bool, 1),
	}, nil
}

func (w *Watcher) AddFile(file string) error {
	return w.fsnotify.Add(file)
}

func (w *Watcher) AddDir(path string) error {
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if err = w.AddFile(path); err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

func (w *Watcher) AddGlob(pattern string) error {
	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := w.AddFile(file); err != nil {
			return err
		}
	}

	return nil
}

func (w *Watcher) On(event fsnotify.Op, f func(string) error) error {
	switch event {
	case fsnotify.Write:
		w.onWrite = f
	case fsnotify.Create:
		w.onCreate = f
	case fsnotify.Remove:
		w.onRemove = f
	case fsnotify.Rename:
		w.onRename = f
	case fsnotify.Chmod:
		w.onChmod = f
	default:
		return fmt.Errorf("event %s not supported", event)
	}

	return nil
}

func (w *Watcher) All(f func(fsnotify.Event, string) error) {
	w.onAll = f
}

func (w *Watcher) Watch() error {
	var err error
	var group errgroup.Group

	if len(w.fsnotify.WatchList()) == 0 {
		return errors.New("no files specified to watch")
	}

	group.Go(func() error {
		for {
			select {
			case event := <-w.fsnotify.Events:
				switch {
				case event.Op&fsnotify.Write == fsnotify.Write:
					if w.onWrite != nil {
						err = w.onWrite(event.Name)
					}
				case event.Op&fsnotify.Create == fsnotify.Create:
					if w.onCreate != nil {
						err = w.onCreate(event.Name)
					}
				case event.Op&fsnotify.Remove == fsnotify.Remove:
					if w.onRemove != nil {
						err = w.onRemove(event.Name)
					}
				case event.Op&fsnotify.Rename == fsnotify.Rename:
					if w.onRename != nil {
						err = w.onRename(event.Name)
					}
				case event.Op&fsnotify.Chmod == fsnotify.Chmod:
					if w.onChmod != nil {
						err = w.onChmod(event.Name)
					}
				}

				if w.onAll != nil {
					err = w.onAll(event, event.Name)
				}

				if err != nil {
					return err
				}

			case <-w.done:
				w.fsnotify.Close()
				close(w.done)
				return nil

			case err := <-w.fsnotify.Errors:
				return err
			}
		}
	})

	if err := group.Wait(); err != nil {
		return err
	}

	return nil
}

func (w *Watcher) Done() {
	w.done <- true
}
