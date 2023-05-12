package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bitfield/script"
	"github.com/fsnotify/fsnotify"
)

const usage = `Usage:

watcher /path/to/file /path/to/dir -- echo "something changed"`

func parseArgs(args []string) ([]string, string, error) {
	paths := []string{}
	exec := []string{}
	isPath := true
	for _, arg := range args {
		if arg == "--" {
			isPath = false
		} else if isPath {
			paths = append(paths, arg)
		} else {
			exec = append(exec, arg)
		}
	}
	if len(paths) < 1 {
		return nil, "", errors.New("No paths to watch")
	}
	if len(exec) < 1 {
		return nil, "", errors.New("No action to perform")
	}
	return paths, parseScript(exec), nil
}

func parseScript(args []string) string {
	exec := []string{args[0]}
	for _, arg := range args[1:] {
		exec = append(exec, escape(arg))
	}
	return strings.Join(exec, " ")
}

func escape(arg string) string {
	arg = strings.ReplaceAll(arg, "'", "'\\''")
	return fmt.Sprintf("'%s'", arg)
}

func main() {
	paths, exec, err := parseArgs(os.Args[1:])
	if err != nil {
		log.Fatalf("Error: %s\n\n%s\n", err.Error(), usage)
	}

	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Watch paths
	for _, path := range paths {
		err := watcher.Add(path)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) {
					log.Println("Modified:", event.Name)
					log.Println("Executing", exec)
					script.Exec(exec).Stdout()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	// Block main goroutine forever.
	<-make(chan struct{})
}
