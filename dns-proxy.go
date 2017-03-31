package main

import (
	"os"
	"fmt"
	"log"
	"net"
	"time"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

var listenPort string = "53"
var resolver string = "192.168.0.1:53"

func resolveAndReply(listener *net.UDPConn, resolverConn *net.UDPConn, addr *net.UDPAddr, buffer []byte, bn int) {
	defer resolverConn.Close()
	responseBuffer := make([]byte, 65535)
	_, err := resolverConn.Write(buffer[0:bn])
	if err != nil {
		log.Println("Error writing to resolver ", err)
		return
	}
	resolverConn.SetReadDeadline(time.Now().Add(8 * time.Second))
	n, err := resolverConn.Read(responseBuffer)
	if err != nil {
		log.Println("Error reading from resolver ", err)
		return
	}
	n, err = listener.WriteToUDP(responseBuffer[0:n], addr)
	if err != nil {
		log.Println("Error replying to original requestor ", err)
	}
	var dnsQ layers.DNS
	var dnsA layers.DNS
	var df gopacket.DecodeFeedback
	var logMessage string = ""
	dnsQ.DecodeFromBytes(buffer, df)
	dnsA.DecodeFromBytes(responseBuffer, df)
	logMessage = logMessage + fmt.Sprintf("%s ", addr.IP)
	for q := range dnsQ.Questions {
		logMessage = logMessage + fmt.Sprintf("%s ", dnsQ.Questions[q].Name)
	}
	for q := range dnsA.Answers {
		if dnsA.Answers[q].IP != nil {
			logMessage = logMessage + fmt.Sprintf("[%s:%s] ", dnsA.Answers[q].Type, dnsA.Answers[q].IP)
		}
		if dnsA.Answers[q].MX.Name != nil {
			logMessage = logMessage + fmt.Sprintf("[%s:%s] ", dnsA.Answers[q].Type, dnsA.Answers[q].MX.Name)
		}
	if dnsA.Answers[q].CNAME != nil {
			logMessage = logMessage + fmt.Sprintf("[%s:%s] ", dnsA.Answers[q].Type, dnsA.Answers[q].CNAME)
		}
	}
	for q := range dnsA.Additionals {
		if dnsA.Additionals[q].IP != nil {
			logMessage = logMessage + fmt.Sprintf("[%s:%s] ", dnsA.Additionals[q].Type, dnsA.Additionals[q].IP)
		}
		if dnsA.Additionals[q].MX.Name != nil {
			logMessage = logMessage + fmt.Sprintf("[%s:%s] ", dnsA.Additionals[q].Type, dnsA.Additionals[q].MX.Name)
		}
		if dnsA.Additionals[q].CNAME != nil {
			logMessage = logMessage + fmt.Sprintf("[%s:%s] ", dnsA.Additionals[q].Type, dnsA.Additionals[q].CNAME)
		}
		if dnsA.Additionals[q].PTR != nil {
			logMessage = logMessage + fmt.Sprintf("[%s:%s] ", dnsA.Additionals[q].Type, dnsA.Additionals[q].PTR)
		}
		if dnsA.Additionals[q].TXT != nil {
			logMessage = logMessage + fmt.Sprintf("[%s:%s] ", dnsA.Additionals[q].Type, dnsA.Additionals[q].TXT)
		}
	}
	for q := range dnsA.Authorities {
		if dnsA.Authorities[q].NS != nil {
			logMessage = logMessage + fmt.Sprintf("[%s:%s] ", dnsA.Authorities[q].Type, dnsA.Authorities[q].NS)
		}
	/*
		if dnsA.Authorities[q].SOA != nil {
			logMessage = logMessage + fmt.Sprintf("[%s:%s] ", dnsA.Authorities[q].Type, dnsA.Authorities[q].SOA)
		}
		if dnsA.Authorities[q].SRV != nil {
			logMessage = logMessage + fmt.Sprintf("[%s:%s] ", dnsA.Authorities[q].Type, dnsA.Authorities[q].SRV)
		}
		if dnsA.Authorities[q].MX != nil {
			logMessage = logMessage + fmt.Sprintf("[%s:%v] ", dnsA.Authorities[q].Type, dnsA.Authorities[q].MX)
		}
	*/
	}
	log.Println(logMessage)
}

func main() {
	log.SetOutput(os.Stdout)
	listenerAddr,err := net.ResolveUDPAddr("udp", ":"+listenPort)
	if err != nil {
		log.Println("Error resolving listener addr.")
		return
	}
	listener, err := net.ListenUDP("udp", listenerAddr)
	if err != nil {
		log.Fatal("Error creating listener ", err)
	}
	defer listener.Close()
	resolverAddr, err := net.ResolveUDPAddr("udp", resolver)
	if err != nil {
		log.Println("Resolving resolver ", err)
		return
	}
	buffer := make([]byte, 65535)
	for {
		n,addr,err := listener.ReadFromUDP(buffer)
		if err != nil {
			log.Println("Error reading UDP port", err)
		}
		defer listener.Close()
		resolverConn, err := net.DialUDP("udp", nil, resolverAddr)
		if err != nil {
			log.Println("Connecting to resolver ", err)
			return
		}
		go resolveAndReply(listener, resolverConn, addr, buffer, n)
	}
}
