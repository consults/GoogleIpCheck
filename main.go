package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"
)

func getIPs() []string {
	url := "https://unpkg.com/@hcfy/google-translate-ip@1.0.0/ips.txt"

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	ips := strings.Split(string(body), "\n")
	return ips
}
func writeHosts(ip string) {
	hostsPath := "/etc/hosts" // Linux/macOS hosts path
	if runtime.GOOS == "windows" {
		hostsPath = "C:\\Windows\\System32\\drivers\\etc\\hosts" // Windows hosts path
	}

	content, err := ioutil.ReadFile(hostsPath)
	if err != nil {
		log.Fatal(err)
	}

	lines := strings.Split(string(content), "\n")
	changeIP := ip + " translate.googleapis.com"

	for index, line := range lines {
		if strings.Contains(line, "translate.googleapis.com") {
			lines[index] = changeIP
		}
	}

	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(hostsPath, []byte(output), 0644)
	if err != nil {
		log.Fatal(err)
	}
}
func checkIP(ip string, q chan string) string {
	HOST := "translate.googleapis.com"
	TESTIP_FORMAT := "https://%s/translate_a/single?client=gtx&sl=en&tl=fr&q=a"
	url := fmt.Sprintf(TESTIP_FORMAT, ip)
	headers := map[string]string{
		"Host": HOST,
	}

	startTime := time.Now()
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("请求错误")
		log.Fatal(err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			// 超时处理
			// fmt.Printf("超时: [%s]\n", ip)
			return ""
		}
	}
	defer func(Body *http.Response) {
		if Body == nil {
			return
		}
		err := Body.Body.Close()
		if err != nil {
			fmt.Println("关闭body失败")
		}
	}(resp)

	useTime := time.Since(startTime)
	fmt.Println("****************")
	fmt.Printf("ip: [%s] | 耗时: [%s]\n", ip, useTime)
	fmt.Println("****************")
	q <- ip
	return ip
}
func main() {
	q := make(chan string)
	ips := getIPs()
	for _, ip := range ips {
		go checkIP(ip, q)
	}
	select {
	case ip := <-q:
		writeHosts(ip)
	case <-time.After(3 * time.Second):
		fmt.Println("超时退出")
	}
}
