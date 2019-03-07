package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

func main() {
	var httpCLient *http.Client

	// //长连接测试
	// var netTransport = &http.Transport{
	// 	Dial: (&net.Dialer{
	// 		Timeout:   10 * time.Second,
	// 		Deadline:  time.Now().Add(3 * time.Second),
	// 		KeepAlive: 10 * time.Second, //非空闲链接的维持时间 链接链接复用，多出连接池的链接，形成ESTABLISHED链接后，KeepAlive时间后client主动关闭了链接，则形成time_wait
	// 	}).Dial,

	// 	TLSHandshakeTimeout:   4 * time.Second,
	// 	ResponseHeaderTimeout: 4 * time.Second,
	// 	ExpectContinueTimeout: 1 * time.Second,

	// 	DisableKeepAlives:   false, //长连接
	// 	MaxIdleConns:        100,
	// 	MaxIdleConnsPerHost: 100,
	// 	IdleConnTimeout: time.Duration(60*2) * time.Second, //表示链接池中闲置的链接，超时时间，如果大于这个时间，则主动关闭，形成time_wait,根据系统time_wait时间后最终释放链接
	// }

	//短连接
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 3 * time.Second,
			//Deadline:  time.Now().Add(1e9),
			//KeepAlive: 5 * time.Second,
		}).Dial,
		// TLSHandshakeTimeout:   4 * time.Second,
		// ResponseHeaderTimeout: 4 * time.Second,
		// ExpectContinueTimeout: 0 * time.Second,

		// MaxIdleConns:          200,
		MaxIdleConnsPerHost: -1,
		DisableKeepAlives:   true, //短连接
	}

	httpCLient = &http.Client{
		Timeout:   time.Second * 5,
		Transport: netTransport,
	}

	// args, err := json.Marshal(struct {
	// 	Name    string `json:"name"`
	// 	Address string `json:"address"`
	// }{"liuhy", "beijing"})
	var wg sync.WaitGroup
	wg.Add(501)

	for i := 0; i < 500; i++ {
		time.Sleep(3)
		go func() {

			valMap := make(map[string]string)
			valMap["name"] = "liuhy"
			valMap["address"] = "beijing"
			args, err := json.Marshal(valMap)

			// args := []byte("name:liuhy,address:beijing")
			ioRead := bytes.NewReader(args)

			//req, err := http.NewRequest("POST", "http://127.0.0.1:8000/hello", ioRead)
			req, err := http.NewRequest("POST", "http://106.37.74.247:8000/hello", ioRead)
			if err != nil {
				log.Fatalf("NewRequest err:%s", err.Error())
			} else {
				req.Header.Set("Content-Type", "application/json")
				//req.Close = true //启动短连接

				res, err := httpCLient.Do(req)
				if err != nil {
					log.Fatalf("httpCLient Do err:%s", err.Error())
				}

				//defer res.Body.Close()
				body, err := ioutil.ReadAll(res.Body)
				if err != nil {
					log.Fatalf("ReadAll Do err:%s", err.Error())
				}
				res.Body.Close()

				//io.Copy(ioutil.Discard,resp.Body)
				fmt.Println(string(body))
			}

			wg.Done()
		}()
	}

	wg.Wait()
}
