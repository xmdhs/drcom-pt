package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"

	"github.com/avast/retry-go/v4"
)

var (
	user      string
	pass      string
	authAddr  string
	command   string
	zeroMKKey string
	version   string
)

func init() {
	flag.StringVar(&user, "u", "", "用户名")
	flag.StringVar(&pass, "p", "", "密码")
	flag.StringVar(&authAddr, "a", "http://172.17.100.100:801/eportal/?c=ACSetting&a=Login&jsVersion=3.0&login_t=2", "认证地址")
	flag.StringVar(&command, "c", "", "登录成功后运行的命令")
	flag.StringVar(&zeroMKKey, "z", "0123456789", "0MKKey 抓包获取")
	flag.StringVar(&version, "v", "1.3.5.201712141.P.W.A", "ver 抓包获取")
	flag.Parse()

	if pass == "" {
		pass = os.Getenv("pass")
	}
}

func main() {
	cxt := context.Background()
	c := &http.Client{Timeout: 5 * time.Second}
	getUrl := get204Url()
	for {
		err := retry.Do(func() error {
			return checkWeb(cxt, c, getUrl())
		}, getRetryOpts(cxt, 5)...)
		if err == nil {
			time.Sleep(5 * time.Second)
			continue
		}
		err = retry.Do(func() error {
			return login(cxt, c, user, pass, zeroMKKey, version, authAddr)
		}, getRetryOpts(cxt, 5)...)
		if err != nil {
			log.Println("登录似乎失败了")
			time.Sleep(10 * time.Second)
		}
		func() {
			cxt, cancel := context.WithTimeout(cxt, 10*time.Second)
			defer cancel()
			c := exec.CommandContext(cxt, "sh", "-c", command)
			c.Run()
		}()
	}
}

func get204Url() func() string {
	i := atomic.Uint64{}
	urls := []string{
		"https://connect.rom.miui.com/generate_204",
		"https://connectivitycheck.platform.hicloud.com/generate_204",
		"https://wifi.vivo.com.cn/generate_204",
		"https://conn1.oppomobile.com/generate_204",
	}
	l := uint64(len(urls))
	return func() string {
		n := i.Add(1)
		a := n % l
		return urls[a]
	}
}

var ErrNot204 = errors.New("不是 204")

func checkWeb(cxt context.Context, c *http.Client, url string) error {
	req, err := http.NewRequestWithContext(cxt, "GET", url, nil)
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
	if reps.StatusCode != 204 {
		return fmt.Errorf("checkWeb: %w", ErrNot204)
	}
	return nil
}

func login(cxt context.Context, c *http.Client, user, pass, zeroMKKey, version, authAddr string) error {
	v := url.Values{}
	v.Set("DDDDD", user)
	v.Set("upass", pass)
	v.Set("0MKKey", zeroMKKey)
	v.Set("ver", version)
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
