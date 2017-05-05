package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"sync"
)

const usage = `usage: blast <url-template> < some-file.txt`

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, usage)
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(1)
	}

	ch := make(chan string)
	defer close(ch)

	var wg sync.WaitGroup
	wg.Add(runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			startBlaster(ch)
		}()
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		ch <- fmt.Sprintf(flag.Arg(0), scanner.Text())
	}

	wg.Wait()

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error reading from stdin:", err)
		os.Exit(1)
	}
}

func startBlaster(ch <-chan string) {
	for s := range ch {
		resp, err := http.Get(s)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			continue
		}
		switch sc := resp.StatusCode; sc {
		case http.StatusOK:
			io.Copy(os.Stdout, resp.Body)
			fmt.Println()
		}
		resp.Body.Close()
	}
}
