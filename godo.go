package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

var (
	PROJECT   string
	IMPORT    string
	VERSION   string
	BUILDTIME string
	PLATFORM  string
	BRANCH    string
	REVISION  string
)

func shortVersion() string {
	return fmt.Sprintf("%s %s", PROJECT, VERSION)
}

func longVersion() string {
	build := BUILDTIME
	if "" != BRANCH && "" != REVISION {
		build = fmt.Sprintf("(%s@%s) %s", BRANCH, REVISION, build)
	}
	return fmt.Sprintf("%s %s %s", shortVersion(), PLATFORM, build)
}

func main() {

	var (
		argVersion bool
		argJobs    uint
		argNoEnv   bool
		argQuiet   bool
	)

	flag.BoolVar(&argVersion, "v", false, "Display version information.")
	flag.UintVar(&argJobs, "j", uint(runtime.NumCPU()), "Process a maximum of `count` jobs simultaneously.")
	flag.BoolVar(&argNoEnv, "e", false, "Jobs do not inherit environment.")
	flag.BoolVar(&argQuiet, "q", false, "Do not use stdout for default job output.")
	flag.Parse()

	if argVersion {
		fmt.Println(longVersion())
	} else {

		// unbuffered channel, so we have to ensure all receivers are ready before we
		// begin sending commands to the channel.
		var work sync.WaitGroup
		queue := make(chan string)

		// spawn worker goroutines to process multiple files simultaneously
		for i := uint(0); i < argJobs; i++ {
			go func(w *sync.WaitGroup, q chan string) {
				ip := regexp.MustCompile(`^(STD)?IN=["']?(.+)["']?$`)
				op := regexp.MustCompile(`^(STD)?OUT=["']?(.+)["']?$`)
				for c := range q {
					f := strings.Fields(c)
					var ie, oe, exe string
					var arg []string
					for j := 0; j < len(f); j++ {
						if m := ip.FindStringSubmatch(f[j]); len(m) > 2 {
							ie = m[2]
						} else if m := op.FindStringSubmatch(f[j]); len(m) > 2 {
							oe = m[2]
						} else {
							exe = f[j]
							if j+1 < len(f) {
								arg = f[j+1:]
							}
							break
						}
					}

					if "" != exe {
						cmd := exec.Command(exe, arg...)
						if !argQuiet {
							cmd.Stdout = os.Stdout
						}
						if "" != oe {
							cmd.Stdout = out([]string{oe})
						}
						if "" != ie {
							cmd.Stdin = in([]string{ie})
						}
						if !argNoEnv {
							cmd.Env = os.Environ()
						}
						cmd.Run()
					}

					w.Done()
				}
			}(&work, queue)
		}

		sc := bufio.NewScanner(in(flag.Args()))
		for sc.Scan() {
			work.Add(1)
			queue <- sc.Text()
		}

		// notify the worker goroutines to clean up, no more commands are coming
		close(queue)
		// ensure all of the worker goroutines have finished
		work.Wait()
	}
}

func out(args []string) io.Writer {
	if 1 == len(args) && args[0] != "" {
		if w, err := os.OpenFile(
			args[0], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); nil == err {
			return w
		}
	}
	return os.Stdout
}

func in(args []string) io.Reader {
	switch len(args) {
	case 0:
		return os.Stdin
	case 1:
		if r, err := os.OpenFile(args[0], os.O_RDONLY, 0); nil == err {
			return r
		}
	}
	return strings.NewReader(strings.Join(args, " "))
}
