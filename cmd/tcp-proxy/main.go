package main

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"

	proxy "github.com/jpillora/go-tcp-proxy"
)

var (
	version = "0.0.0-src"
	matchid = uint64(0)
	connid  = uint64(0)
	logger  proxy.ColorLogger

	//localAddr   = flag.String("l", ":9999", "local address")
	//remoteAddr  = flag.String("r", "localhost:80", "remote address")
	//verbose     = flag.Bool("v", false, "display server actions")
	//veryverbose = flag.Bool("vv", false, "display server actions and all tcp data")
	//nagles      = flag.Bool("n", false, "disable nagles algorithm")
	//hex         = flag.Bool("h", false, "output hex")
	//colors      = flag.Bool("c", false, "output ansi colors")
	//unwrapTLS   = flag.Bool("unwrap-tls", false, "remote connection with TLS exposed unencrypted locally")
	//match       = flag.String("match", "", "match regex (in the form 'regex')")
	//replace     = flag.String("replace", "", "replace regex (in the form 'regex~replacer')")
)

func main() {
	//flag.Parse()

	config := proxy.ReadConfig()
	settings := config.Settings
	proxyLink := config.ProxyLinks[0]

	logger := proxy.ColorLogger{
		Verbose: settings.Verbose,
		Color:   settings.OutputAnsiColors,
	}

	logger.Info("go-tcp-proxy (%s) proxing from %v to %v ", version, proxyLink.LocalAddr, proxyLink.RemoteAddr)

	laddr, err := net.ResolveTCPAddr("tcp", proxyLink.LocalAddr)
	if err != nil {
		logger.Warn("Failed to resolve local address: %s", err)
		os.Exit(1)
	}
	raddr, err := net.ResolveTCPAddr("tcp", proxyLink.RemoteAddr)
	if err != nil {
		logger.Warn("Failed to resolve remote address: %s", err)
		os.Exit(1)
	}
	listener, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		logger.Warn("Failed to open local port to listen: %s", err)
		os.Exit(1)
	}

	matcher := createMatcher(settings.MatchRegex)
	replacer := createReplacer(settings.ReplaceRegex)

	if settings.VeryVerbose {
		settings.Verbose = true
	}

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			logger.Warn("Failed to accept connection '%s'", err)
			continue
		}
		connid++

		var p *proxy.Proxy
		if settings.UnwrapTls {
			logger.Info("Unwrapping TLS")
			p = proxy.NewTLSUnwrapped(conn, laddr, raddr, proxyLink.RemoteAddr)
		} else {
			p = proxy.New(conn, laddr, raddr)
		}

		p.Matcher = matcher
		p.Replacer = replacer

		p.Nagles = settings.DisableNaglesAlgorithm
		p.OutputHex = settings.OutputHex
		p.Log = proxy.ColorLogger{
			Verbose:     settings.Verbose,
			VeryVerbose: settings.VeryVerbose,
			Prefix:      fmt.Sprintf("Connection #%03d ", connid),
			Color:       settings.OutputAnsiColors,
		}

		go p.Start()
	}
}

func createMatcher(match string) func([]byte) {
	if match == "" {
		return nil
	}
	re, err := regexp.Compile(match)
	if err != nil {
		logger.Warn("Invalid match regex: %s", err)
		return nil
	}

	logger.Info("Matching %s", re.String())
	return func(input []byte) {
		ms := re.FindAll(input, -1)
		for _, m := range ms {
			matchid++
			logger.Info("Match #%d: %s", matchid, string(m))
		}
	}
}

func createReplacer(replace string) func([]byte) []byte {
	if replace == "" {
		return nil
	}
	//split by / (TODO: allow slash escapes)
	parts := strings.Split(replace, "~")
	if len(parts) != 2 {
		logger.Warn("Invalid replace option")
		return nil
	}

	re, err := regexp.Compile(string(parts[0]))
	if err != nil {
		logger.Warn("Invalid replace regex: %s", err)
		return nil
	}

	repl := []byte(parts[1])

	logger.Info("Replacing %s with %s", re.String(), repl)
	return func(input []byte) []byte {
		return re.ReplaceAll(input, repl)
	}
}
