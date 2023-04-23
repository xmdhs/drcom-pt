package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"github.com/avast/retry-go/v4"
)

var (
	user      string
	pass      string
	authAddr  string
	command   string
	zeroMKKey string
)

func init() {
	flag.StringVar(&user, "u", "", "")
	flag.StringVar(&pass, "p", "", "")
	flag.StringVar(&authAddr, "a", "http://172.17.100.100:801/eportal/?c=ACSetting&a=Login&jsVersion=3.0&login_t=2", "")
	flag.StringVar(&command, "c", "", "")
	flag.StringVar(&zeroMKKey, "z", "0123456789", "")
	flag.Parse()
}

func main() {
	cxt := context.Background()
	c := &http.Client{Timeout: 5 * time.Second}
	for {
		err := retry.Do(func() error {
			return checkWeb(cxt, c)
		}, getRetryOpts(cxt, 5)...)
		if err == nil {
			time.Sleep(5 * time.Second)
			continue
		}
		err = retry.Do(func() error {
			return login(cxt, c, user, pass, zeroMKKey, authAddr)
		}, getRetryOpts(cxt, 5)...)
		if err != nil {
			panic(err)
		}
		func() {
			cxt, cancel := context.WithTimeout(cxt, 10*time.Second)
			defer cancel()
			c := exec.CommandContext(cxt, "sh", "-c", command)
			c.Run()
		}()
	}
}

func checkWeb(cxt context.Context, c *http.Client) error {
	req, err := http.NewRequestWithContext(cxt, "GET", "https://connect.rom.miui.com/generate_204", nil)
	if err != nil {
		return fmt.Errorf("checkWeb: %w", err)
	}
	reps, err := c.Do(req)
	if reps != nil {
		defer reps.Body.Close()
	}
	if err != nil {
		return fmt.Errorf("checkWeb: %w", err)
	}
	_, err = io.Copy(io.Discard, reps.Body)
	if err != nil {
		return fmt.Errorf("checkWeb: %w", err)
	}
	return nil
}

func login(cxt context.Context, c *http.Client, user, pass, zeroMKKey, authAddr string) error {
	v := url.Values{}
	v.Set("DDDDD", user)
	v.Set("upass", pass)
	v.Set("0MKKey", zeroMKKey)
	v.Set("ver", "1.3.5.201712141.P.W.A")
	req, err := http.NewRequestWithContext(cxt, "POST", authAddr, strings.NewReader(v.Encode()))
	if err != nil {
		return fmt.Errorf("login: %w", err)
	}
	req.Header.Set("User-Agent", "DrCOM-HTTPCLIENT")
	req.Header.Set("Accept-Language", "zh-CN")
	req.Header.Set("Charset", "utf-8")
	req.Header.Set("Connection", "Close")

	resp, err := c.Do(req)
	if resp != nil {
		defer resp.Body.Close()

	}
	if err != nil {
		return fmt.Errorf("login: %w", err)
	}
	return nil
}

func getRetryOpts(cxt context.Context, attempts uint) []retry.Option {
	return []retry.Option{
		retry.Attempts(attempts),
		retry.LastErrorOnly(true),
		retry.MaxDelay(20 * time.Minute),
		retry.Context(cxt),
		retry.OnRetry(func(n uint, err error) {
			log.Printf("retry %d: %v", n, err)
		}),
	}
}
