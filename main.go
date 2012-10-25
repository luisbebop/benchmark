package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net"
	"time"
	"sync/atomic"
)

func panicOnErr(msg string, err error) {
	if err!=nil {
		panic(msg+": "+err.Error())
	}
}

func main() {
	var counter uint64 = 0
	done := make(chan bool, 1)
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
				//connect to the socket
				conn, err := net.Dial("tcp", *host)
				if err!=nil {
					fmt.Println("dial:", err)
					return
				}
				
				//more one socket connected
				atomic.AddUint64(&counter, 1)
				
				//write data to the socket
				if nr, err := conn.Write(data); err!=nil || nr!=len(data) {
					log.Fatalln("Unable to write", err)
				}
				
				//receive from the socket
				reply := make([]byte, len(data))
				if nr, err := conn.Read(reply); err!=nil || nr!=len(data) {
					log.Fatalln("Unable to read", err)
				}

				//check error on receive
				panicOnErr("copy", err)
				
				//compare response from the socket
				if bytes.Compare(reply, data)!=0 {
					log.Fatalln("This is not an echo server: "+string(reply))
				}
				
				//sleep 1 second for each req to the server
				time.Sleep(time.Second * 1)
			}
		}()
	}
	
	go func() {
		var diffTransactions uint64 = 0
		totalTransactions := (uint64)((*nthreads) * (*nreqs))
		for {
			//get the transaction counter every 1 second
			counterPartial := atomic.LoadUint64(&counter)
			
			//print to stdout
		    fmt.Printf("Transaction %v/%v tps:%v\n", counterPartial, totalTransactions, counterPartial - diffTransactions)
			
			//check if all transactions were proccessed
			if counterPartial == totalTransactions {
				done<-true
			}
			
			//store the last counter to know how much new transactions were proccessed in the last second
			diffTransactions = counterPartial
			
			//sleep 1 second to refresh the stdout
			time.Sleep(time.Second * 1)
		}
	}()
	
	<-done
	fmt.Println("benchmark finished")
}