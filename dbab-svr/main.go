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
	"net/http"
	"os"
	"strings"

	echo "github.com/labstack/echo/v4"
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

func pixelServ(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("ETag", "dbab")
	w.Header().Set("Connection", "close")
	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	w.Write([]byte(pixel))
}

func (h *proxyHandler) handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Connection", "close")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write([]byte(h.setting))
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
	autoProxy := fmt.Sprintf(
		"function FindProxyForURL(url, host) { return \"PROXY %s:3128; DIRECT\"; }",
		readFile(proxyFile))
	autoProxyBuf := []byte(autoProxy)

	e := echo.New()

	serveAutoProxy := func(c echo.Context) error {
		response := c.Response()
		response.Header().Add("Connection", "close")
		return c.Blob(http.StatusOK, "application/octet-stream", autoProxyBuf)
	}
	e.GET("/proxy.pac", serveAutoProxy)
	e.GET("/wpad.dat", serveAutoProxy)

	pixelBuf := []byte(pixel)
	servePixel := func(c echo.Context) error {
		response := c.Response()
		response.Header().Add("Cache-Control", "public, max-age=31536000")
		response.Header().Add("Connection", "close")
		response.Header().Add("ETag", "dbab")
		return c.Blob(http.StatusOK, "image/gif", pixelBuf)
	}
	e.GET("*", servePixel)

	log.Printf("starting dbab pixel server on port %s\n", httpPort)
	// Start server
	e.Logger.Print(e.Start(httpPort))
	log.Fatal("dbab pixel server stopped.")
}
