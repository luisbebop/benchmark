package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net"
	"time"
)

func panicOnErr(msg string, err error) {
	if err!=nil {
		panic(msg+": "+err.Error())
	}
}

func main() {
	nthreads := flag.Int("n", -1, "how many threads will open connection concurrently")
	msg := flag.String("data", "hello", "what will those threads write")
	host := flag.String("host", "localhost:1234", "all threads connect to host")
	nreqs := flag.Int("req",1, "how many req per thread")
	flag.Parse()
	
	if *nthreads<=0 {
		fmt.Println("Must supply nthreads")
		flag.Usage()
		return
	}
	
	fmt.Println("starting benchmark")
	data := []byte(*msg+"\n")
	for i:=0; i<*nthreads; i++ {
		go func() {
			for j:=0; j<*nreqs; j++ {
				conn, err := net.Dial("tcp", *host)
				if err!=nil {
					fmt.Println("dial:", err)
					return
				}
				if nr, err := conn.Write(data); err!=nil || nr!=len(data) {
					log.Fatalln("Unable to write", err)
				}
				reply := make([]byte, len(data))
				if nr, err := conn.Read(reply); err!=nil || nr!=len(data) {
					log.Fatalln("Unable to read", err)
				}
				
				panicOnErr("copy", err)
				
				if bytes.Compare(reply, data)!=0 {
					log.Fatalln("This is not an echo server: "+string(reply))
				}
				
				time.Sleep(time.Second * 1)
			}
		}()
	}
	fmt.Println("benchmark finished")
	select {}
}