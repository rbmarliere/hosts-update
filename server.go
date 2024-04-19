// SPDX-License-Identifier: GPL-2.0

package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
)

var local_secret = "TEST"

type custom_logger struct {
	verbose bool
	*log.Logger
}

var logger = &custom_logger{false, log.New(os.Stdout, "", 0)}

func (l *custom_logger) Println(v ...interface{}) {
	if l.verbose {
		l.Logger.Println(v...)
	}
}

func parseHeaders(r *http.Request) ([]string, error) {
	var msg string

	secret := r.Header.Get("Secret")
	if secret == "" {
		msg = "Secret not found"
		logger.Println(msg)
		return nil, errors.New(msg)
	}
	logger.Println("Found Secret: " + secret)

	ip := r.Header.Get("IP")
	if ip == "" {
		msg = "IP not found"
		logger.Println(msg)
		return nil, errors.New(msg)
	}
	logger.Println("Found IP: " + ip)

	name := r.Header.Get("Name")
	if name == "" {
		msg = "Name not found"
		logger.Println(msg)
		return nil, errors.New(msg)
	}
	logger.Println("Found Name: " + name)

	return []string{secret, ip, name}, nil
}

func updateHost(ip string, name string) {
	pattern := fmt.Sprintf(`(?m)^.*%s$`, name)
	logger.Println("Pattern: ", pattern)
	re := regexp.MustCompile(pattern)

	hosts, err := os.ReadFile("/etc/hosts")
	if err != nil {
		logger.Println("Error reading hosts file: ", err)
		return
	}

	host := fmt.Sprintf("%s %s", ip, name)
	matches := re.FindAllSubmatch([]byte(hosts), -1)
	if len(matches) > 0 {
		logger.Println("Found name.")
		new_hosts := re.ReplaceAllString(string(hosts), host)
		err := os.WriteFile("/etc/hosts", []byte(new_hosts), 0644)
		if err != nil {
			logger.Println("Error writing hosts file: ", err)
		}
	} else {
		logger.Println("No matches.")
		hosts, err := os.OpenFile("/etc/hosts", os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			logger.Println("Error opening hosts file: ", err)
			return
		}
		host := fmt.Sprintf("%s %s\n", ip, name)
		_, err = hosts.WriteString(host)
		if err != nil {
			logger.Println("Error appending hosts file: ", err)
		}
		hosts.Close()
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	headers, error := parseHeaders(r)
	if error != nil {
		w.WriteHeader(http.StatusBadRequest)
		logger.Println("Bad Request")
		return
	}

	secret := headers[0]
	ip := headers[1]
	name := headers[2]

	if secret == local_secret {
		updateHost(ip, name)
		w.WriteHeader(http.StatusOK)
		logger.Println("OK")
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		logger.Println("Unauthorized")
	}
}

func main() {
	port := flag.String("port", "32888", "port to listen on")
	verbose := flag.Bool("verbose", false, "verbose logging")
	flag.Parse()

	logger.verbose = *verbose

	http.HandleFunc("/", handleRequest)
	http.ListenAndServe(":"+*port, nil)
}
