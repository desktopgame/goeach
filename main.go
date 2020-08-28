package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/buildkite/interpolate"
	"github.com/desktopgame/filetree"
)

func main() {
	var (
		cmd      = flag.String("cmd", "echo ${file}", "execution command")
		dir      = flag.String("dir", ".", "target directory")
		info     = flag.Bool("i", false, "show a command, but not execute command")
		pattern  = flag.String("p", ".+", "pattern for target file")
		parallel = flag.Bool("a", false, "execution in parallel by goroutine")
	)
	flag.Parse()
	if *info && *parallel {
		fmt.Println("* -a option is ignored *")
	}
	node, err := filetree.CollectLimited(*dir, nil, 42)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	wg := &sync.WaitGroup{}
	files := node.Flatten()
	regex := regexp.MustCompile(*pattern)
	for _, file := range files {
		if file.Name == "." {
			continue
		}
		if !regex.MatchString(file.Name) {
			continue
		}
		env := interpolate.NewSliceEnv([]string{
			"file=" + file.Name,
			"path=" + file.Path,
			"dir=" + filepath.Dir(file.Path),
		})
		formattedCmd, _ := interpolate.Interpolate(env, *cmd)
		if err != nil {
			continue
		}
		if *info {
			fmt.Println(formattedCmd)
			continue
		}
		args := strings.Split(formattedCmd, " ")
		cmd := exec.Command(args[0], args[1:]...)
		if *parallel {
			wg.Add(1)
			go func() {
				fmt.Println(formattedCmd)
				cmd.Run()
				wg.Done()
			}()
		} else {
			fmt.Println(formattedCmd)
			cmd.Run()
		}
	}
	wg.Wait()
}
