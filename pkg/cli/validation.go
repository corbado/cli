package cli

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func (c *CLI) validateLocalAddress(localAddress string) string {
	parsedURL, err := url.Parse(localAddress)
	if err != nil {
		return err.Error()
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "must have http or https scheme"
	}

	if len(parsedURL.Path) > 0 {
		return "must have empty path"
	}

	if len(parsedURL.RawQuery) > 0 {
		return "must have no query parameters"
	}

	if len(parsedURL.Fragment) > 0 {
		return "must have no fragment"
	}

	if parsedURL.User != nil {
		return "must have no user authentication"
	}

	port := "80"
	if parsedURL.Port() != "" {
		port = parsedURL.Port()
	}

	host := fmt.Sprintf("%s:%s", parsedURL.Hostname(), port)

	if _, err := net.DialTimeout("tcp", host, time.Second*3); err != nil {
		return fmt.Sprintf("%s not reachable", host)
	}

	return ""
}

func (c *CLI) validateProjectID(id string) bool {
	if !strings.HasPrefix(id, "pro-") {
		return false
	}

	if _, err := strconv.ParseUint(id[4:], 10, 64); err != nil {
		return false
	}

	return true
}
