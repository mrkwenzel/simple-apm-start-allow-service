package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

var requests = make(map[string]string)

var cookies = make(map[string]time.Time)

func issueCookie(response http.ResponseWriter, request *http.Request) {
	ttl := 5 * time.Minute
	expire := time.Now().Add(ttl)
	val, _ := randomHex(32)
	cookie := http.Cookie{
		Name:    "MRHSession",
		Value:   val,
		Expires: expire,
	}
	http.SetCookie(response, &cookie)
	requests[cookie.Value] = request.URL.Path
	cookies[cookie.Value] = expire
	http.Redirect(response, request, "/my.policy", http.StatusFound)
}

func proxyHandler(response http.ResponseWriter, request *http.Request) {
	cookie, err := request.Cookie("MRHSession")
	if err == http.ErrNoCookie {
		issueCookie(response, request)
	} else if err != nil {
		fmt.Println(err)
		fmt.Fprintf(response, "Something went terribly wrong... %s", err)
	} else {
		sub := cookies[cookie.Value].Sub(time.Now())
		fmt.Println(sub)
		if sub <= 0 {
			issueCookie(response, request)
		} else {
			serveProxy("http://localhost:8081", response, request)
		}
	}
}

func myPolicyHandler(response http.ResponseWriter, request *http.Request) {
	cookie, err := request.Cookie("MRHSession")
	if request.Method != http.MethodGet {
		http.Redirect(response, request, "/vdesk/hangup.php3", http.StatusFound)
	} else if err == http.ErrNoCookie {
		http.Redirect(response, request, "/", http.StatusFound)
	} else {
		http.Redirect(response, request, requests[cookie.Value], http.StatusFound)
	}
}

func hangupHandler(response http.ResponseWriter, request *http.Request) {
	fmt.Fprint(response, "Evil page...")
}

func serveProxy(target string, response http.ResponseWriter, request *http.Request) {
	url, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ServeHTTP(response, request)
}

func randomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func main() {
	http.HandleFunc("/", proxyHandler)
	http.HandleFunc("/my.policy", myPolicyHandler)
	http.HandleFunc("/vdesk/hangup.php3", hangupHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
