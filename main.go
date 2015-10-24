package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

var (
	catflag = flag.NewFlagSet("esact", flag.ExitOnError)
	urlStr  = catflag.String("url", "http://127.0.0.1:9200", "Elasticsearch instance locator")
	help    = catflag.Bool("help", false, "Output its available columns")
	verbose = catflag.Bool("verbose", false, "Turn on verbose output")
)

func cat(cmd string) ([]byte, error) {
	u, err := url.Parse(*urlStr)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "_cat", cmd)

	q := u.Query()
	if *verbose {
		q.Set("v", "true")
	}
	if *help {
		q.Set("help", "true")
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func catHelp() {
	b, _ := cat("")
	list := strings.Fields(string(b))
	printed := []string{}
	for _, v := range list {
		f := true
		for _, s := range printed {
			if strings.HasPrefix(v, s) {
				f = false
			}
		}
		if f && strings.HasPrefix(v, "/_cat") {
			fmt.Printf("    %s\n", strings.TrimPrefix(v, "/_cat/"))
			printed = append(printed, v)
		}
	}
}

func main() {
	flag.Usage = func() {
		flag.PrintDefaults()
		fmt.Println(`Usage of escat:
  Options:
    --help    Output its available columns
    --verbose Turn on verbose output`)
		fmt.Println(`
  Commands:`)
		catHelp()
	}
	flag.Parse()
	os.Exit(run())
}

func run() int {
	const (
		success  = 0
		failure  = 1
		notfound = 127
	)

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		return success
	}
	name := args[0]
	if name == "help" {
		flag.Usage()
		return success
	}
	s := 1
	if len(args) > 1 {
		if other := args[1]; !strings.HasPrefix(other, "-") {
			name = path.Join(name, other)
		}
		s = 2
	}
	catflag.Parse(args[s:])

	b, err := cat(name)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return failure
	}

	fmt.Fprintln(os.Stdout, string(b))

	return success
}
