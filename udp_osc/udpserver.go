package main

import (
	"bitbucket.org/liamstask/gosc"
	"flag"
	"fmt"
	"github.com/augustoroman/hexdump"
	"github.com/gobs/cmd"
	"log"
	"net"
	"strings"
)

var port = flag.Int("port", 12345, "Port to listen on.")

var msg, addr string

func main() {
	flag.Parse()
	fmt.Println("Welcome!")

	server := newServer(*port)
	go server.run()

	commander := &cmd.Cmd{
		EnableShell: true,
	}
	commander.Init()
	commander.Add(cmd.Command{
		Name: "quit",
		Help: "Quit",
		Call: func(string) bool { return true },
	})
	commander.Add(cmd.Command{
		Name: "send",
		Help: "Send [addr] [text]",
		Call: func(txt string) bool { server.Send(parse(txt)); return false },
	})
	commander.Add(cmd.Command{
		Name: "hex",
		Help: "hex true/false",
		Call: func(txt string) bool { server.hex = (txt == "true"); return false },
	})

	oscaddr := fmt.Sprintf(":%d", *port+1)
	go func() {
		for {
			log.Println("Listening for OSC on", oscaddr)
			err := osc.ListenAndServeUDP(oscaddr, server)
			log.Println("osc server exited: ", err)
		}
	}()

	commander.CmdLoop()
}

func parse(txt string) (addr, text string) {
	splits := strings.SplitN(txt, " ", 2)
	return splits[0], splits[1]
}

type UdpServer struct {
	conn *net.UDPConn
	hex  bool
}

func (u *UdpServer) Dispatch(b *osc.Bundle) {
	log.Println("Got OSC message: ", *b)
	log.Printf("  Got %d messages:", len(b.Messages))
	for i, msg := range b.Messages {
		log.Printf("  %d: %s %d %v", i, msg.Address, len(msg.Args), msg.Args)
	}
}

func newServer(port int) *UdpServer {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: port})
	if err != nil {
		panic(err)
	}

	return &UdpServer{
		conn: conn,
	}
}

func (u *UdpServer) quit() {
	u.conn.Close()
}

func (u *UdpServer) run() {
	log.Println("Listening on", u.conn.LocalAddr())
	const maxPacketSize = 1 << 20
	data := make([]byte, maxPacketSize)
	for {
		msglen, from, err := u.conn.ReadFromUDP(data)
		if err != nil {
			log.Println("Udp error: ", err)
			break
		}
		var strdata string
		if u.hex {
			strdata = hexdump.Dump(data[:msglen])
		} else {
			strdata = string(data[:msglen])
		}
		log.Printf("Got %d bytes from %v: %s", msglen, from, strdata)
	}
}

func (u *UdpServer) Send(addr, text string) {
	a, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Printf("Failed to resolve addr %s: %v", addr, err)
		return
	}
	_, err = u.conn.WriteToUDP([]byte(text), a)
	if err != nil {
		log.Printf("Failed to send to %s: %v", addr, err)
		return
	}

}
