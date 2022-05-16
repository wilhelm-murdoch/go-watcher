package watcher

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"golang.org/x/sync/errgroup"
)

// Watcher represents a wrapper around `fsnotify` complete with support for its
// own callbacks for all supported event types.
type Watcher struct {
	fsnotify                                       *fsnotify.Watcher                              // Instance of `fsnotify` wrapped by this package.
	done                                           chan bool                                      // A signal channel used to exit the wait loop.
	onRemove, onCreate, onWrite, onRename, onChmod func(os.FileInfo, error) error                 // Dedicated optional callback functions for each specific `fsnotify.Event` type.
	onAll                                          func(fsnotify.Event, os.FileInfo, error) error // Dedicated optional callback function used to catch all `fsnotify.Event` types.
}

// New creates a new instance of a Watcher struct.
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

// AddFile adds a single valid file to the current Watcher instance and returns
// an error if the file is not valid.
func (w *Watcher) AddFile(path string) error {
	return w.fsnotify.Add(path)
}

// AddDir will recursively walk the specified directory tree and add all valid
// files to the current watcher instance for monitoring.
func (w *Watcher) AddDir(path string) error {
	err := filepath.WalkDir(path, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			if err = w.AddFile(path); err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

// AddGlob will monitor the specified "glob" pattern and add all valid files to
// the current watcher instance for monitoring.
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

// On fires off an assigned callback for each event type. Only specified events
// are supported and all will return either nil or an error. Every watcher
// instances exits when it first encounters an error.
func (w *Watcher) On(event fsnotify.Op, f func(os.FileInfo, error) error) error {
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

// All will fire off the specified callback on any supported `fsnotify` event.
func (w *Watcher) All(f func(fsnotify.Event, os.FileInfo, error) error) {
	w.onAll = f
}

// Watch creates a new `errgroup` instance and monitors for changes to any of
// the specified files. All supported event types will fire off specified
// callbacks if available. This method exits on the first encountered error.
func (w *Watcher) Watch() error {
	var group errgroup.Group

	if len(w.fsnotify.WatchList()) == 0 {
		return errors.New("no files specified to watch")
	}

	if w.onAll == nil &&
		w.onWrite == nil &&
		w.onCreate == nil &&
		w.onRemove == nil &&
		w.onRename == nil &&
		w.onChmod == nil {
		return errors.New("no event type callbacks have been defined; nothing to process")
	}

	group.Go(func() error {
		for {
			select {
			case event := <-w.fsnotify.Events:
				info, err := os.Stat(event.Name)
				switch {
				case event.Op&fsnotify.Write == fsnotify.Write:
					if w.onWrite != nil {
						err = w.onWrite(info, err)
					}
				case event.Op&fsnotify.Create == fsnotify.Create:
					if w.onCreate != nil {
						err = w.onCreate(info, err)
					}
				case event.Op&fsnotify.Remove == fsnotify.Remove:
					if w.onRemove != nil {
						err = w.onRemove(info, err)
					}
				case event.Op&fsnotify.Rename == fsnotify.Rename:
					if w.onRename != nil {
						err = w.onRename(info, err)
					}
				case event.Op&fsnotify.Chmod == fsnotify.Chmod:
					if w.onChmod != nil {
						err = w.onChmod(info, err)
					}
				}

				if w.onAll != nil {
					err = w.onAll(event, info, err)
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

	return group.Wait()
}

// Done signals a blocking channel that processing is complete and that we can
// safely exit the current watcher instance.
func (w *Watcher) Done() {
	w.done <- true
}
