////////////////////////////////////////////////////////////////////////////
// Program: dbab-svr
// Purpose: Pixel Server in Go
// Authors: Tong Sun (c) 2019, All rights reserved
////////////////////////////////////////////////////////////////////////////

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
	"time"
)

////////////////////////////////////////////////////////////////////////////
// Constant and data type/structure definitions

const (
	confFile  = "/etc/dbab/dbab.addr"
	proxyFile = "/etc/dbab/dbab.proxy"

	pixel = "\x47\x49\x46\x38\x39\x61\x01\x00\x01\x00\x80\x00\x00\xFF\xFF\xFF\x00\x00\x00\x21\xF9\x04\x01\x00\x00\x00\x00\x2C\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x02\x44\x01\x00\x3B"
)

////////////////////////////////////////////////////////////////////////////
// Global variables definitions

var (
	progname = "dbab-svr"
)

////////////////////////////////////////////////////////////////////////////
// Function definitions

//==========================================================================
// support functions

// readFile returns the single-line file content (less trailing \n) of the file by the given fname
func readFile(fname string) string {
	b, err := ioutil.ReadFile(fname)
	abortOn("Reading input file: "+fname, err)
	return strings.TrimSuffix(string(b), "\n")
}

// abortOn will quit on anticipated errors gracefully without stack trace
func abortOn(errCase string, e error) {
	if e != nil {
		fmt.Fprintf(os.Stderr, "[%s] %s error: %v\n", progname, errCase, e)
		os.Exit(1)
	}
}

//==========================================================================
// Main

func main() {
	httpPort := readFile(confFile)
	_httpPort := os.Getenv("DBAB_SERVER_ADDR")
	if _httpPort != "" {
		httpPort = _httpPort

	}
	autoProxy := fmt.Sprintf(
		"function FindProxyForURL(url, host) { return \"PROXY %s:3128; DIRECT\"; }",
		readFile(proxyFile))

	log.Printf("starting dbab pixel server on port %s\n", httpPort)
	l, err := net.Listen("tcp", httpPort)
	if err != nil {
		log.Fatal(err)
	}
	re := regexp.MustCompile(`\s*(\w+)\s*([^\s]+)\s*HTTP\/(\d.\d)`)
	for {
		c, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		if err := c.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
			log.Fatal(err)
		}
		s := bufio.NewScanner(c)
		var req struct {
			Method  string
			URL     string
			Version string
		}
		for s.Scan() {
			line := s.Text()
			matches := re.FindStringSubmatch(line)
			if matches != nil {
				req.Method = strings.ToUpper(matches[1])
				req.URL = matches[2]
				req.Version = matches[3]
				continue
			}
			if line == "" {
				break
			}
		}
		if err := s.Err(); err != nil {
			continue
		}
		if req.Method == "GET" && (req.URL == "/proxy.pac" || req.URL == "/wpad.dat") {
			fmt.Fprintf(c, "HTTP/1.0 200 OK\r\n")
			fmt.Fprintf(c, "Connection: close\r\n")
			fmt.Fprintf(c, "Content-Type: application/octet-stream\r\n\r\n")
			c.Write([]byte(autoProxy))
		} else {
			fmt.Fprintf(c, "HTTP/1.0 200 OK\r\n")
			fmt.Fprintf(c, "ETag: dbab\r\n")
			fmt.Fprintf(c, "Connection: close\r\n")
			fmt.Fprintf(c, "Cache-Control: public, max-age=31536000\r\n")
			fmt.Fprintf(c, "Content-type: image/gif\r\n")
			fmt.Fprintf(c, "Content-length: 43\r\n\r\n")
			c.Write([]byte(pixel))
		}
		c.Close()
	}
	l.Close()
	log.Fatal("dbab pixel server stopped.")
}
