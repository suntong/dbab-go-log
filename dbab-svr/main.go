////////////////////////////////////////////////////////////////////////////
// Program: dbab-svr
// Purpose: Pixel Server in Go
// Authors: Tong Sun (c) 2019, All rights reserved
////////////////////////////////////////////////////////////////////////////

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/valyala/fasthttp"
)

////////////////////////////////////////////////////////////////////////////
// Constant and data type/structure definitions

const (
	confFile  = "/etc/dbab/dbab.addr"
	proxyFile = "/etc/dbab/dbab.proxy"

	pixel = "\x47\x49\x46\x38\x39\x61\x01\x00\x01\x00\x80\x00\x00\xFF\xFF\xFF\x00\x00\x00\x21\xF9\x04\x01\x00\x00\x00\x00\x2C\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x02\x44\x01\x00\x3B"
)

type proxyHandler struct {
	setting string
}

////////////////////////////////////////////////////////////////////////////
// Global variables definitions

var (
	progname = "dbab-svr"
)

////////////////////////////////////////////////////////////////////////////
// Function definitions

func pixelServ(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("ETag", "dbab")
	ctx.Response.Header.Set("Connection", "close")
	ctx.Response.Header.Set("Content-Type", "image/gif")
	ctx.Response.Header.Set("Cache-Control", "public, max-age=31536000")
	fmt.Fprintf(ctx, "%s", pixel)
}

func (h *proxyHandler) handler(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Connection", "close")
	ctx.Response.Header.Set("Content-Type", "application/octet-stream")
	fmt.Fprintf(ctx, "%s", h.setting)
}

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
	autoProxy := &proxyHandler{fmt.Sprintf(
		"function FindProxyForURL(url, host) { return \"PROXY %s:3128; DIRECT\"; }",
		readFile(proxyFile))}

	m := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/proxy.pac", "/wpad.dat":
			autoProxy.handler(ctx)
		default:
			pixelServ(ctx)
		}
	}
	log.Printf("starting dbab pixel server on port %s\n", httpPort)
	// Run the web server.
	log.Print(fasthttp.ListenAndServe(httpPort, m))
	log.Fatal("dbab pixel server stopped.")
}
