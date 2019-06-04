////////////////////////////////////////////////////////////////////////////
// Program: dbab-svr
// Purpose: Pixel Server dbab-svr in Go
// Authors: Tong Sun (c) 2019, All rights reserved
////////////////////////////////////////////////////////////////////////////

package main

import (
	"log"
	"net/http"
	"os"
)

////////////////////////////////////////////////////////////////////////////
// Constant and data type/structure definitions

const (
	pixel = "\x47\x49\x46\x38\x39\x61\x01\x00\x01\x00\x80\x00\x00\xFF\xFF\xFF\x00\x00\x00\x21\xF9\x04\x01\x00\x00\x00\x00\x2C\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x02\x44\x01\x00\x3B"
)

////////////////////////////////////////////////////////////////////////////
// Global variables definitions

var (
	httpPort = ":80" // default is :80
)

////////////////////////////////////////////////////////////////////////////
// Function definitions

func pixelserv(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("ETag", "dbab")
	w.Header().Set("Connection", "close")
	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	w.Write([]byte(pixel))
}

//==========================================================================
// Main

func main() {
	_httpPort := os.Getenv("DBAB_SERVER_ADDR")
	if _httpPort != "" {
		httpPort = _httpPort
	}

	http.HandleFunc("/", pixelserv)
	log.Printf("Starting dbab pixel server on port %s\n", httpPort)
	// Run the web server.
	log.Fatal(http.ListenAndServe(httpPort, nil))
}
