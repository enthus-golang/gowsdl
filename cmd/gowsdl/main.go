// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
/*

Gowsdl generates Go code from a WSDL file.

This project is originally intended to generate Go clients for WS-* services.

Usage: gowsdl [options] myservice.wsdl
  -o string
        File where the generated code will be saved (default "myservice.go")
  -p string
        Package under which code will be generated (default "myservice")
  -v    Shows gowsdl version

Features

Supports only Document/Literal wrapped services, which are WS-I (http://ws-i.org/) compliant.

Attempts to generate idiomatic Go code as much as possible.

Supports WSDL 1.1, XML Schema 1.0, SOAP 1.1.

Resolves external XML Schemas

Supports providing WSDL HTTP URL as well as a local WSDL file.

Not supported

UDDI.

TODO

Add support for filters to allow the user to change the generated code.

If WSDL file is local, resolve external XML schemas locally too instead of failing due to not having a URL to download them from.

Resolve XSD element references.

Support for generating namespaces.

Make code generation agnostic so generating code to other programming languages is feasible through plugins.

*/

package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"time"

	gen "github.com/hooklift/gowsdl"
)

// Version is initialized in compilation time by go build.
var Version string

// Name is initialized in compilation time by go build.
var Name string

var vers = flag.Bool("v", false, "Shows gowsdl version")
var pkg = flag.String("p", "myservice", "Package under which code will be generated")
var outFile = flag.String("o", "myservice.go", "File where the generated code will be saved")
var dir = flag.String("d", "./", "Directory under which package directory will be created")
var insecure = flag.Bool("i", false, "Skips TLS Verification")
var makePublic = flag.Bool("make-public", true, "Make the generated types public/exported")

// HTTP client configuration flags
var httpTimeout = flag.Duration("http-timeout", 30*time.Second, "Timeout for HTTP requests")
var httpRetries = flag.Int("http-retries", 0, "Number of retries for failed requests")
var httpRetryDelay = flag.Duration("http-retry-delay", 1*time.Second, "Delay between retries")
var httpRateLimit = flag.Int("http-rate-limit", 0, "Rate limit for HTTP requests per second (0 = no limit)")
var httpUserAgent = flag.String("http-user-agent", "gowsdl/1.0", "User-Agent header for requests")
var httpProxy = flag.String("http-proxy", "", "HTTP proxy URL")
var tlsCACert = flag.String("tls-ca-cert", "", "Path to CA certificate file")
var tlsClientCert = flag.String("tls-client-cert", "", "Path to client certificate file (requires -tls-client-key)")
var tlsClientKey = flag.String("tls-client-key", "", "Path to client key file (requires -tls-client-cert)")
var tlsMinVersion = flag.String("tls-min-version", "1.2", "Minimum TLS version (1.0, 1.1, 1.2, 1.3)")

func init() {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)
	log.SetPrefix("🍀  ")
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] myservice.wsdl\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	// Show app version
	if *vers {
		log.Println(Version)
		os.Exit(0)
	}

	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(0)
	}

	wsdlPath := os.Args[len(os.Args)-1]

	if *outFile == wsdlPath {
		log.Fatalln("Output file cannot be the same WSDL file")
	}

	// Create HTTP client configuration
	httpConfig := gen.DefaultHTTPClientConfig()
	httpConfig.Timeout = *httpTimeout
	httpConfig.MaxRetries = *httpRetries
	httpConfig.RetryDelay = *httpRetryDelay
	httpConfig.RateLimitPerSecond = *httpRateLimit
	httpConfig.UserAgent = *httpUserAgent
	httpConfig.ProxyURL = *httpProxy
	httpConfig.InsecureSkipVerify = *insecure

	// Configure TLS
	if *tlsMinVersion != "" {
		var minVersion uint16
		switch *tlsMinVersion {
		case "1.0":
			minVersion = tls.VersionTLS10
		case "1.1":
			minVersion = tls.VersionTLS11
		case "1.2":
			minVersion = tls.VersionTLS12
		case "1.3":
			minVersion = tls.VersionTLS13
		default:
			log.Fatalf("Invalid TLS version: %s", *tlsMinVersion)
		}
		if httpConfig.TLSConfig == nil {
			httpConfig.TLSConfig = &tls.Config{}
		}
		httpConfig.TLSConfig.MinVersion = minVersion
	}

	// Add CA certificate if provided
	if *tlsCACert != "" {
		if _, err := httpConfig.WithCACert(*tlsCACert); err != nil {
			log.Fatalf("Failed to load CA certificate: %v", err)
		}
	}

	// Add client certificate if provided
	if *tlsClientCert != "" && *tlsClientKey != "" {
		if _, err := httpConfig.WithClientCert(*tlsClientCert, *tlsClientKey); err != nil {
			log.Fatalf("Failed to load client certificate: %v", err)
		}
	} else if *tlsClientCert != "" || *tlsClientKey != "" {
		log.Fatalln("Both -tls-client-cert and -tls-client-key must be provided together")
	}

	// load wsdl
	gowsdl, err := gen.NewGoWSDLWithConfig(wsdlPath, *pkg, httpConfig, *makePublic)
	if err != nil {
		log.Fatalln(err)
	}

	// generate code
	gocode, err := gowsdl.Start()
	if err != nil {
		log.Fatalln(err)
	}

	pkg := filepath.Join(*dir, *pkg)
	err = os.Mkdir(pkg, 0744)

	file, err := os.Create(filepath.Join(pkg, *outFile))
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	data := new(bytes.Buffer)
	data.Write(gocode["header"])
	data.Write(gocode["types"])
	data.Write(gocode["operations"])
	data.Write(gocode["soap"])

	// go fmt the generated code
	source, err := format.Source(data.Bytes())
	if err != nil {
		file.Write(data.Bytes())
		log.Fatalln(err)
	}

	file.Write(source)

	// server
	serverFile, err := os.Create(pkg + "/" + "server" + *outFile)
	if err != nil {
		log.Fatalln(err)
	}
	defer serverFile.Close()

	serverData := new(bytes.Buffer)
	serverData.Write(gocode["server_header"])
	serverData.Write(gocode["server_wsdl"])
	serverData.Write(gocode["server"])

	serverSource, err := format.Source(serverData.Bytes())
	if err != nil {
		serverFile.Write(serverData.Bytes())
		log.Fatalln(err)
	}
	serverFile.Write(serverSource)

	log.Println("Done 👍")
}
