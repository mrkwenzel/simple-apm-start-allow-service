package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"time"
)

// Service port
var port = "8080"

// Time to live for the MRHSession cookie
var cookieTTL = "360"

// Proxied service
var proxiedService = "http://localhost:8081"

/*
	Request map to redirect to the right ressource after issuing the MRHSession
	cookie.
*/
var requests = make(map[string]string)

// Issued cookies together with their expiration time
var cookies = make(map[string]time.Time)

/*
	Redirect with issuing the cookie
	It also stores the originally requested ressource in the "requests" map
	and the expiration time of the issued cookie in the "cookies" map.
*/
func issueCookie(response http.ResponseWriter, request *http.Request) {
	ttl, err := strconv.Atoi(cookieTTL)
	if err != nil {
		fmt.Printf("Something went terribly wrong...")
		log.Fatal(err)
	}
	expire := time.Now().Add(time.Duration(ttl) * time.Second)
	val, _ := randomHex(32)
	cookie := http.Cookie{
		Name:    "MRHSession",
		Value:   val,
		Expires: expire,
		Path:    "/",
	}
	http.SetCookie(response, &cookie)
	requests[cookie.Value] = request.URL.Path // Store requested ressource
	cookies[cookie.Value] = expire            // Store expiration of MRHSession cookie
	http.Redirect(response, request, "/my.policy", http.StatusFound)
}

// Proxy handler for imitating the APM flow
func proxyHandler(response http.ResponseWriter, request *http.Request) {
	cookie, err := request.Cookie("MRHSession")
	if err == http.ErrNoCookie {
		issueCookie(response, request) // No cookie -> Issue new one
	} else if err != nil {
		fmt.Println(err) // hopefully never happens...
		fmt.Fprintf(response, "Something went terribly wrong... %s", err)
	} else {
		sub := cookies[cookie.Value].Sub(time.Now())
		if sub <= 0 {
			// If cookie is expired -> Issue new one
			issueCookie(response, request)
		} else {
			// Proxy the request
			url, _ := url.Parse(proxiedService)
			proxy := httputil.NewSingleHostReverseProxy(url)
			proxy.ServeHTTP(response, request)
		}
	}
}

// Handler for imitating 302 responses of the APM
func myPolicyHandler(response http.ResponseWriter, request *http.Request) {
	cookie, err := request.Cookie("MRHSession")
	if request.Method != http.MethodGet {
		// If the request method is not a GET request -> Evil!
		http.Redirect(response, request, "/vdesk/hangup.php3", http.StatusFound)
	} else if err == http.ErrNoCookie {
		// This should never happen, but if so -> Start from the beginning
		http.Redirect(response, request, "/", http.StatusFound)
	} else {
		// Whoop! This is the lucky way! Let's start with the cookie.
		http.Redirect(response, request, requests[cookie.Value], http.StatusFound)
	}
}

// Handler for evil hangup page
func hangupHandler(response http.ResponseWriter, request *http.Request) {
	fmt.Fprint(response, "Evil page...")
}

// Utility function for creating something similar to a real MRHSession uuid
func randomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func main() {
	envPort := os.Getenv("SIMPLE_APM_PORT")
	if envPort != "" {
		port = envPort
	}

	envProxiedService := os.Getenv("SIMPLE_APM_PROXIED_SERVICE")
	if envProxiedService != "" {
		proxiedService = envProxiedService
	}

	envCookieTTL := os.Getenv("SIMPLE_APM_COOKIE_TTL")
	if envCookieTTL != "" {
		cookieTTL = envCookieTTL
	}

	fmt.Printf("Listening on: http://localhost:%s\n", port)
	fmt.Printf("Proxied service: %s\n", proxiedService)
	fmt.Printf("Cookie ttl set to: %s seconds\n", cookieTTL)

	http.HandleFunc("/", proxyHandler)
	http.HandleFunc("/my.policy", myPolicyHandler)
	http.HandleFunc("/vdesk/hangup.php3", hangupHandler)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
