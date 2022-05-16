package watcher_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wilhelm-murdoch/go-watcher"
)

func TestWatcherNoCallbacks(t *testing.T) {
	var err error
	w, err := watcher.New()
	assert.Nil(t, err, "was expecting no errors, but got %s instead", err)

	err = w.AddFile("test.txt")
	assert.Nil(t, err, "was expecting no errors, but got %s instead", err)

	err = w.Watch()
	assert.NotNil(t, err, "was expecting no errors, but got %s instead", err)
}

func TestWatcherNoFiles(t *testing.T) {
	var err error
	w, err := watcher.New()
	assert.Nil(t, err, "was expecting no errors, but got %s instead", err)

	err = w.Watch()
	assert.NotNil(t, err, "was expecting no errors, but got %s instead", err)
}

func TestWatcherWalkDir(t *testing.T) {

}

func TestWatcherGlob(t *testing.T) {

}

func TestWatcherOn(t *testing.T) {

}

func TestWatcherAl(t *testing.T) {

}

func TestWatcherWatch(t *testing.T) {

}

func TestWatcherDone(t *testing.T) {

}
