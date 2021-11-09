package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	version = "0.0.1"
)

type arrayHeaders []string

func (i *arrayHeaders) String() string {
	return "my string representation"
}

func (i *arrayHeaders) Set(value string) error {
	*i = append(*i, value)
	return nil
}

type (
	Config struct {
		Url      string
		Headers  arrayHeaders
		Silent   bool
		Json     bool
		UseStdin bool
		Threads  int
		Timeout  int
	}

	Target struct {
		Host   string
		Method string
	}

	Output struct {
		Code   int    `json:"code"`
		Target string `json:"target"`
		Method string `json:"method"`
		Length int    `json:"length"`
	}
)

var (
	config         Config
	methodPayloads = []string{
		"CHECKIN",
		"CHECKOUT",
		"CONNECT",
		"COPY",
		"DELETE",
		"GET",
		"HEAD",
		"INDEX",
		"LINK",
		"LOCK",
		"MKCOL",
		"MOVE",
		"NOEXISTE",
		"OPTIONS",
		"ORDERPATCH",
		"PATCH",
		"POST",
		"PROPFIND",
		"PROPPATCH",
		"PUT",
		"REPORT",
		"SEARCH",
		"SHOWMETHOD",
		"SPACEJUMP",
		"TEXTSEARCH",
		"TRACE",
		"TRACK",
		"UNCHECKOUT",
		"UNLINK",
		"UNLOCK",
		"VERSION-CONTROL",
	}
	targetList []*url.URL
	stdout     *log.Logger
)

func main() {
	log.SetFlags(0)
	stdout = log.New(os.Stdout, "", 0)
	parseArguments()
	if !config.Silent {
		showBanner()
	}

	targets := make(chan Target, 0)
	wg := sync.WaitGroup{}
	wg.Add(config.Threads)

	for config.Threads > 0 {
		config.Threads -= 1
		go func() {
			for {
				target := <-targets
				if target.Host == "" {
					break
				}
				checkTarget(target)
			}
			wg.Done()
		}()
	}

	for _, t := range targetList {
		us := t.String()
		for _, m := range methodPayloads {
			targets <- Target{Host: us, Method: m}
		}
	}

	close(targets)
	wg.Wait()

}

func checkTarget(target Target) {
	code, length, err := Request(target, time.Duration(config.Timeout)*time.Second)
	if err != nil {
		return
	}
	var res string

	if config.Json {
		v := Output{Code: code, Target: target.Host, Method: target.Method, Length: length}
		o, err := json.Marshal(v)
		if err != nil {
			return
		}
		res = string(o)
	} else {
		space := strings.Repeat(" ", 17 - len(target.Method))
		res = fmt.Sprintf("%v\t%v\t%v%v%v", code, length, target.Method, space, target.Host)
	}
	stdout.Println(res)
}

func showBanner() {
	log.Println("\n██╗░░░██╗███████╗██████╗░██████╗░░░░░░░████████╗░█████╗░███╗░░░███╗██████╗░███████╗██████╗░\n██║░░░██║██╔════╝██╔══██╗██╔══██╗░░░░░░╚══██╔══╝██╔══██╗████╗░████║██╔══██╗██╔════╝██╔══██╗\n╚██╗░██╔╝█████╗░░██████╔╝██████╦╝█████╗░░░██║░░░███████║██╔████╔██║██████╔╝█████╗░░██████╔╝\n░╚████╔╝░██╔══╝░░██╔══██╗██╔══██╗╚════╝░░░██║░░░██╔══██║██║╚██╔╝██║██╔═══╝░██╔══╝░░██╔══██╗\n░░╚██╔╝░░███████╗██║░░██║██████╦╝░░░░░░░░░██║░░░██║░░██║██║░╚═╝░██║██║░░░░░███████╗██║░░██║\n░░░╚═╝░░░╚══════╝╚═╝░░╚═╝╚═════╝░░░░░░░░░░╚═╝░░░╚═╝░░╚═╝╚═╝░░░░░╚═╝╚═╝░░░░░╚══════╝╚═╝░░╚═╝", version)
	if !config.Json {
		log.Println("Code\tLength\tMethod\t\t target")
	}
}

func parseArguments() {
	flag.BoolVar(&config.Silent, "silent", false, "Disable banner")
	flag.BoolVar(&config.Json, "json", false, "Output format as json")
	flag.StringVar(&config.Url, "url", "", "comma separated Urls to check")
	flag.IntVar(&config.Threads, "t", 10, "Number of threads to use")
	flag.IntVar(&config.Timeout, "timeout", 10, "Seconds to wait before timeout.")
	flag.BoolVar(&config.UseStdin, "stdin", false, "Read targets url from stdin")
	flag.Var(&config.Headers, "h", "request headers ex: -h 'X-Api-Key: MyApiKey' -h 'Content-Type: application/json'")
	flag.Parse()

	targetList = make([]*url.URL, 0)
	if !config.UseStdin {
		urls := strings.Split(config.Url, ",")
		for _, us := range urls {
			ok, u := isValidUrl(us)
			if ok {
				targetList = append(targetList, u)
			}
		}
	} else {
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			ok, u := isValidUrl(sc.Text())
			if ok {
				targetList = append(targetList, u)
			}
		}
		if err := sc.Err(); err != nil {
			log.Fatalln(err)
		}
	}

	if len(targetList) == 0 {
		log.Println("error: empty target list")
		flag.Usage()
		os.Exit(1)
	}
}

func isValidUrl(toTest string) (bool, *url.URL) {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false, nil
	}

	u, err := url.Parse(toTest)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false, nil
	}

	return true, u
}
