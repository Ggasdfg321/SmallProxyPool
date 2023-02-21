package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"
)

type cmd struct {
	command string
}

var (
	c = cmd{}
)

func (c cmd) Show(args []string) {
	if len(args) > 1 {
		if args[1] == "ip" {
			for _, i := range alive_address_time {
				if i[0] == tmp_addr {
					defer func() {
						if err := recover(); err != nil {
							fmt.Println("当前使用的IP地址是", tmp_addr, i[2], "延迟", "错误")
							return
						}
					}()
					socksProxy := "socks5://" + tmp_addr
					proxy := func(_ *http.Request) (*url.URL, error) {
						return url.Parse(socksProxy)
					}
					tr := &http.Transport{
						TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
						Proxy:           proxy,
					}
					start := time.Now()
					url := "https://opendata.baidu.com/api.php"
					req, _ := http.NewRequest("GET", url, nil)
					req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36")
					(&http.Client{Transport: tr, Timeout: 8 * time.Second}).Do(req)
					end := time.Since(start)
					fmt.Println("当前使用的IP地址是", tmp_addr, i[2], "延迟", end)
					return
				}
			}
			fmt.Println("请先使用代理后在执行这条命令")
			return
		} else if args[1] == "all" {
			fmt.Println("正在输出全部IP地址")
			fmt.Println("---------------------------------")
			for _, i := range alive_address_time {
				fmt.Println(i[0], i[1], i[2])
			}
			fmt.Println("---------------------------------")
			fmt.Println("一共有",len(alive_address_time),"代理")
			return
		}
	}

}

func (c cmd) Use(args []string) {
	if len(args) > 1 {
		if strings.Contains(args[1], ":") {
			setProxy = true
			useProxy = args[1]
			fmt.Println("设置代理",useProxy,"成功")
			return
		} else if args[1] == "random" {
			setProxy = false
			useProxy = ""
			fmt.Println("设置random成功")
			return
		}
	}
	fmt.Println(args[0], ":参数错误！")
}

func Command() {

	for {
		if len(alive_address) > 0 {
			c.command = ""
			defer func() {
				if err := recover(); err != "" {
					fmt.Println(err)
					fmt.Println(c.command+":", "指令错误")
					Command()
				}
			}()
			fmt.Print("-> ")
			// fmt.Scanln(&c.command)
			reader := bufio.NewReader(os.Stdin)
			c.command, _ = reader.ReadString('\n')
			c.command = strings.TrimSpace(c.command)
			if c.command == "" {
				continue
			}
			funcs := reflect.ValueOf(&c)
			comm := strings.ToUpper(c.command[:1]) + c.command[1:]
			var args []reflect.Value
			if len(strings.Split(c.command, " ")) > 1 {
				comm = strings.Split(comm, " ")[0]
				args = []reflect.Value{reflect.ValueOf(strings.Split(c.command, " "))}
			} else {
				args = []reflect.Value{reflect.ValueOf([]string{c.command})}
			}
			funcs.MethodByName(comm).Call(args)
		}

	}
}
