package integration_test

import "testing"

// The names net/http gives a status code, an HTTP method, and a connection
// state are the whole numbers Go gives them, so a program that names one gets
// the number it got in Go, and a phrase for a code the table of phrases holds
// is the phrase Go returns, with the empty phrase for a code it does not.
func TestHTTPStatusMethodAndStateConstants(t *testing.T) {
	runParityTest(t, `package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println(http.StatusOK, http.StatusNotFound, http.StatusInternalServerError, http.StatusContinue)
	fmt.Println(http.StatusNetworkAuthenticationRequired, http.StatusTeapot, http.StatusIMUsed)
	fmt.Println(http.StatusRequestHeaderFieldsTooLarge, http.StatusEarlyHints, http.StatusMovedPermanently)
	fmt.Println(http.StatusPermanentRedirect, http.StatusNonAuthoritativeInfo, http.StatusUnavailableForLegalReasons)
	fmt.Println(http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodHead)
	fmt.Println(http.MethodPatch, http.MethodOptions, http.MethodConnect, http.MethodTrace)
	fmt.Println(int(http.StateNew), int(http.StateActive), int(http.StateIdle), int(http.StateHijacked), int(http.StateClosed))
	fmt.Println(http.DefaultMaxHeaderBytes, http.DefaultMaxIdleConnsPerHost)
	fmt.Println(http.StatusText(200), http.StatusText(418), http.StatusText(999))
	fmt.Println(http.StatusText(511), http.StatusText(103), http.StatusText(226))
}
`)
}
