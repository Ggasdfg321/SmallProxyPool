package main

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"main/balance"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-ini/ini"
	"github.com/panjf2000/ants/v2"
	"github.com/valyala/fastjson"
)

var (
	conf       *ini.File
	fofa_email string
	fofa_key   string
	thread     int
)

var (
	address            []string
	alive_address      []string
	alive_address_time [][]string
	b                  = &balance.RoundRobinBalance{}
	wg                 sync.WaitGroup
)

var (
	tmp      int
	tmp_try  int
	tmp_addr string
)

var (
	setProxy = false
	useProxy = ""
)

func init() {
	fmt.Println("作者：Ggasdfg321 By T00ls.Com")
	tcfgs, err := ini.Load("config.ini")
	if err != nil {
		tconf, _ := base64.StdEncoding.DecodeString("W2dsb2JhbF0KZW1haWwgPSAKa2V5ID0gCnJ1bGUgPSAncHJvdG9jb2w9PSJzb2NrczUiICYmICJWZXJzaW9uOjUgTWV0aG9kOk5vIEF1dGhlbnRpY2F0aW9uKDB4MDApIiAmJiBjb3VudHJ5PSJDTiInCmJpbmRfaXAgPSAxMjcuMC4wLjEKYmluZF9wb3J0ID0gMTA4MA==")
		ioutil.WriteFile("config.ini", tconf, 0666)
	}
	conf = tcfgs
	fofa_email = conf.Section("global").Key("email").Value()
	fofa_key = conf.Section("global").Key("key").Value()
	thread, _ = strconv.Atoi(conf.Section("global").Key("thread").Value())
	tmp_try = 0
}
func main() {
	fmt.Println("[*]", "正在获取socks5代理中")
	go getSocks5Data()
	go checkAlive()
	go Command()
	add := conf.Section("global").Key("bind_ip").Value() + ":" + conf.Section("global").Key("bind_port").Value()
	server, err := net.Listen("tcp", add)
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		client, err := server.Accept()
		if err != nil {
			fmt.Printf("Accept failed :%v", err)
			continue
		}
		go process(client)
	}
}

func getSocks5Data() error {

	tmp = 1
	for {
		if tmp > 1 {
			time.Sleep(60 * time.Second)
		}
		tmp = tmp + 1
		req, err := http.NewRequest("GET", "https://fofa.info/api/v1/search/all", nil)
		if err != nil {
			fmt.Println(err)
			return err
		}
		tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		// rule := `protocol=="socks5" && "Version:5 Method:No Authentication(0x00)" && country="CN"`
		rule := conf.Section("global").Key("rule").Value()
		rule = base64.StdEncoding.EncodeToString([]byte(rule))
		r := req.URL.Query()
		r.Add("email", fofa_email)
		r.Add("key", fofa_key)
		r.Add("qbase64", rule)
		r.Add("size", "2000")
		req.URL.RawQuery = r.Encode()
		req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36")
		resp, err := (&http.Client{Transport: tr}).Do(req)
		if err != nil {
			fmt.Println(err)
			return err
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		var p fastjson.Parser
		v, _ := p.Parse(string(body))
		if v.GetStringBytes("errmsg") != nil {
			fmt.Println(string(body))
			return err
		}
		var rst []string
		for _, i := range v.GetArray("results") {
			ipaddr := string(i.GetStringBytes("1")) + ":" + string(i.GetStringBytes("2"))
			rst = append(rst, ipaddr)
		}
		address = rst
		// fmt.Println("获取成功，总查询数量：", len(address))
	}
}

func checkAlive() {
	lock := &sync.Mutex{}
	var fastj fastjson.Parser
	for {
		if len(address) == 0 {
			time.Sleep(1 * time.Second)
		}
		// fmt.Println("[*]", "正在过滤代理中")
		p, _ := ants.NewPoolWithFunc(thread, func(i interface{}) {
			socks5 := i.(string)
			defer func() {
				if err := recover(); err != nil {
					if slicesFind(alive_address, socks5) == true {
						alive_address = sliceDelete(alive_address, socks5).([]string)
						alive_address_time = sliceDelete(alive_address_time, socks5).([][]string)
					}
					wg.Done()
					return
				}
			}()
			socksProxy := "socks5://" + socks5
			proxy := func(_ *http.Request) (*url.URL, error) {
				return url.Parse(socksProxy)
			}
			tr := &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				Proxy:           proxy,
			}
			url := fmt.Sprintf("https://opendata.baidu.com/api.php?query=%s&co=&resource_id=6006", strings.Split(socks5, ":")[0])
			start := time.Now().UnixNano()
			req, _ := http.NewRequest("GET", url, nil)
			req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36")
			req.Header.Add("Connection", "close")
			resp, _ := (&http.Client{Transport: tr, Timeout: 8 * time.Second}).Do(req)
			stop := time.Now().UnixNano() - start
			if resp.StatusCode == 200 {
				defer resp.Body.Close()
				body, _ := ioutil.ReadAll(resp.Body)
				v, _ := fastj.Parse(string(body))
				location := ConvertByte2String(v.GetStringBytes("data", "0", "location"), "GB18030")
				if location != "" {
					if slicesFind(alive_address, socks5) == false {
						sliceAlive := []string{socks5, strconv.FormatInt(stop, 10), location}
						lock.Lock()
						alive_address = append(alive_address, socks5)
						alive_address_time = append(alive_address_time, sliceAlive)
						lock.Unlock()
					}
				}

			}
			wg.Done()
		})
		defer p.Release()
		for _, i := range address {
			wg.Add(1)
			_ = p.Invoke(i)
		}
		wg.Wait()
		b.Set(alive_address)
		// fmt.Println("[*]", "过滤完成，一共有", len(alive_address), "/", len(alive_address_time), "个代理可用")
		time.Sleep(10 * time.Second)

	}
}

func process(client net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			process(client)
		}
	}()
	if len(alive_address) == 0 {
		time.Sleep(1 * time.Second)
		process(client)
	}
	defer client.Close()
	addr := getproxy(tmp_try)
	tmp_addr = addr
	// fmt.Println("当前用的ip是",addr)
	cc, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		fmt.Println("connect error:", err)
		tmp_try = 1 + tmp_try
		process(client)
	}
	tmp_try = 0
	defer cc.Close()
	go io.Copy(cc, client)
	io.Copy(client, cc)
}
func getproxy(arg int) string {
	polling := conf.Section("rule").Key("polling").Value()
	polling = strings.ToLower(polling)
	if setProxy == true {
		return useProxy
	} else {
		if polling == "true" {
			return b.Get()
		} else {
			var times []int64
			for _, i := range alive_address_time {
				t, err := strconv.ParseInt(i[1], 10, 64)
				if err != nil {
					fmt.Println("[-]", "代理", alive_address_time[0], "时间转换失败")
					return ""
				}
				times = append(times, t)
				sort.Slice(times, func(i, j int) bool {
					return times[i] < times[j]
				})
			}
			one := times[0]
			two := times[1]
			three := times[2]
			for _, i := range alive_address_time {
				t, _ := strconv.ParseInt(i[1], 10, 64)
				switch arg {
				case 0:
					if t == one {
						return i[0]
					}
				case 1:
					if t == two {
						return i[0]
					}
				case 2:
					if t == three {
						return i[0]
					}
				default:
					if t == one {
						return i[0]
					}
				}

			}
			fmt.Println("[-]", "当前无代理")
			return ""
		}
	}
}
func printAddr() {
	for {
		if tmp_addr == "" {
			time.Sleep(5 * time.Second)
			printAddr()
		}
		fmt.Println("[+]", "当前使用的代理IP是", tmp_addr)
		time.Sleep(8 * time.Second)
	}

}
