# Watcher

A no-nonsense wrapper around the [fsnotify](https://pkg.go.dev/github.com/fsnotify/fsnotify) package.

## Install
```
$ go get github.com/wilhelm-murdoch/go-watcher
```

## Usage

Watch for new files in `/path/to/files` and stop the watcher on the first `fsnotify.Create` event.

```go
package main 

import (
  "fmt"

  "github.com/wilhelm-murdoch/go-watcher"
)

func main() {
  w, err := watcher.New()

  w.AddDir("/path/to/files")

  w.On(fsnotify.Create, func(file os.FileInfo, err error) error {
    fmt.Println("new file:", file.Name())
    w.Done()
    return nil
  })

  w.Watch()
}
```