package main

import (
	"http"
	"flag"
	"log"
	"net"
	"fmt"
	"os"
	"regexp"
	"io"
	"io/ioutil"
)

func main() {
	flag.Parse()

	http.Handle("/", http.HandlerFunc(WriteResponse))
	//http.Handle("/", http.HandlerFunc(TestCode))
	//http.Handle("/", http.HandlerFunc(TestCode2))
	log.Stdout("start server")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Exit("ListenAndServe: ", err.String())
	}
}

func TestCode2(c *http.Conn, req *http.Request) {
	dump, err := http.DumpRequest(req, true)
	if err != nil {
		io.WriteString(c, err.String())
		return
	}
	_, err = c.Write(dump)
	if err != nil {
		io.WriteString(c, err.String())
		return
	}
	log.Stdout(string(dump))
}

func TestCode(c *http.Conn, req *http.Request) {
	PrintRequest(req)
	io.WriteString(c, "test code")
}

func WriteResponse(c *http.Conn, req *http.Request) {

	PrintRequest(req)
	req = DeleteHopByHopHeader(req)
	PrintRequest(req)
	host := req.URL.Host
	if match, _ := regexp.MatchString(":[0-9]+$", host); !match {
		if req.URL.Scheme == "http" {
			host = host + ":80"
		} else if req.URL.Scheme == "https" {
			host = host + ":443"
		}
	}
	proxy := os.Getenv("HTTP_PROXY")
	if len(proxy) > 0 {
		proxy_url, _ := http.ParseURL(proxy)
		host = proxy_url.Host
	}
	log.Stdoutf("host:%s\n", host)

	tcp, err := net.Dial("tcp", "", host)
	if err != nil {
		log.Stdout(err)
		return
	}
	defer tcp.Close()

	conn := http.NewClientConn(tcp, nil)
	defer conn.Close()

	err = conn.Write(req)
	if err != nil {
		log.Stdout(err)
		return
	}
	response, err := conn.Read()
	if err != nil {
		log.Stdout(err)
		return
	}
	
	PrintResponse(response)
	/***
	size := response.ContentLength
	if size > 0 {
		buf := make([]byte, size)
		io.ReadFull(response.Body, buf)
		c.Write(buf)
	} else if response.RequestMethod == "HEAD" {
		buf, _ := ioutil.ReadAll(response.Body)
		c.Write(buf)
	}
	***/
	err = response.Write(c)
	if err != nil {
		log.Stdout("ERROR:")
		log.Stdout(err)
		return
	}
	err = response.Body.Close()
	if err != nil {
		log.Stdout("ERROR:")
		log.Stdout(err)
	}
}

func PrintRequest(r *http.Request) {
	
	fmt.Println("")
	fmt.Println("☆ Reqest-----------")
    fmt.Printf("Proto:%s\n", r.Proto)
    fmt.Printf("Method:%s\n", r.Method)
	fmt.Printf("RawURL:%s\n", r.RawURL)
	fmt.Printf("Host:%s\n", r.Host)
	fmt.Printf("Referer:%s\n", r.Referer)
	fmt.Printf("UserAgent:%s\n", r.UserAgent)
    fmt.Println("TransferEncoding:")
	for encode := range r.TransferEncoding {
		fmt.Printf("\t%s\n", encode)
	}
    fmt.Println("Header:")
    for key, header := range r.Header {
        fmt.Printf("\t%s:%s\n", key, header)
    }
    fmt.Println("Body:")
	size := r.ContentLength
	if size > 0 {
		buf := make([]byte, size)
		io.ReadFull(r.Body, buf)
		fmt.Println(string(buf))
	} else if r.Method == "HEAD" {
		buf, _ := ioutil.ReadAll(r.Body)
		fmt.Println(string(buf))
	}
    fmt.Println("Form:")
    for key, form := range r.Form {
        fmt.Printf("\t%s:%s\n", key, form)
    }
}

func PrintResponse(r *http.Response) {

	fmt.Println("")
	fmt.Println("☆ Response-----------")
    fmt.Printf("Status:%s\n", r.Status)
    fmt.Printf("Proto:%s\n", r.Proto)
    fmt.Printf("RequestMethod:%s\n", r.RequestMethod)
    fmt.Printf("TransferEncoding:\n")
	for encode := range r.TransferEncoding {
		fmt.Printf("\t%s\n", encode)
	}
    fmt.Printf("Header:\n")
    for key, header := range r.Header {
        fmt.Printf("\t%s:%s\n", key, header)
    }
    fmt.Printf("Body:\n")
	size := r.ContentLength
	if size > 0 {
		buf := make([]byte, size)
		io.ReadFull(r.Body, buf)
		fmt.Println(string(buf))
	} else if r.RequestMethod == "HEAD" {
		buf, _ := ioutil.ReadAll(r.Body)
		fmt.Println(string(buf))
	}
}

func DeleteHopByHopHeader(req *http.Request) (*http.Request) {
	var ok bool
	if _, ok = req.Header["Connection"]; ok {
		req.Header["Connection"] = "", false
	}
	if _, ok = req.Header["Keep-Alive"]; ok {
		req.Header["Keep-Alive"] = "", false
	}
	if _, ok = req.Header["Proxy-Authenticate"]; ok {
		req.Header["Proxy-Authenticate"] = "", false
	}
	if _, ok = req.Header["Proxy-Authorization"]; ok {
		req.Header["Proxy-Authorization"] = "", false
	}
	if _, ok = req.Header["TE"]; ok {
		req.Header["TE"] = "", false
	}
	if _, ok = req.Header["Transfer-Encoding"]; ok {
		req.Header["Transfer-Encoding"] = "", false
	}
	if _, ok = req.Header["Upgrade"]; ok {
		req.Header["Upgrade"] = "", false
	}
	
	return req
}

