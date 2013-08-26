package main

import (
	"bitbucket.org/liamstask/gosc"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/augustoroman/misc/chunker"
	"github.com/gobs/cmd"
	"log"
	"net"
	"runtime"
	"strings"
)

var port = flag.Int("port", 12345, "Port to listen on.")

var msg, addr string

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	runtime.GOMAXPROCS(max(runtime.NumCPU(), 4))

	flag.Parse()
	fmt.Println("Welcome!")

	server := newServer(*port)
	go Process(server.Packets(), server.BufferReturn())
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

func Process(packets <-chan []byte, emptyreturn chan<- []byte) {
	var md chunker.MessageDechunker
	for data := range packets {
		process(&md, data)
		emptyreturn <- data[:cap(data)]
	}
}

func process(md *chunker.MessageDechunker, data []byte) {
	chunk, err := chunker.ParseChunk(data)
	if err != nil {
		log.Println("Malformed chunk: %v", err)
		return
	}

	log.Printf("Message %d, Chunk %d/%d: %d bytes",
		chunk.MessageId(), chunk.ChunkId(), chunk.NumChunks(),
		len(chunk.Data()))

	if md.AddChunk(chunk) {
		msg := md.Assemble()
		log.Printf("Got complete message of %d bytes\n", len(msg))
		var parsed interface{}
		err := json.Unmarshal([]byte(msg), &parsed)
		if err != nil {
			log.Printf("  [Failed to parse message to json: %v]", err)
			log.Println("  Raw message: ", msg)
		} else {
			s, _ := json.MarshalIndent(parsed, "  ", "  ")
			log.Println(string(s))
		}
	}

}

func parse(txt string) (addr, text string) {
	splits := strings.SplitN(txt, " ", 2)
	return splits[0], splits[1]
}

type UdpServer struct {
	databuffers chan []byte
	conn        *net.UDPConn
	hex         bool
	packets     chan []byte
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

	const NUM_BUFFERS = 1000
	const BUFFER_SIZE = 64 << 10

	buffers := make(chan []byte, NUM_BUFFERS)
	for i := 0; i < NUM_BUFFERS; i++ {
		buffers <- make([]byte, BUFFER_SIZE)
	}

	return &UdpServer{
		conn:        conn,
		databuffers: buffers,
		packets:     make(chan []byte, NUM_BUFFERS),
	}
}

func (u *UdpServer) quit() {
	close(u.packets)
	u.conn.Close()
}

func (u *UdpServer) Packets() <-chan []byte {
	return u.packets
}

func (u *UdpServer) BufferReturn() chan<- []byte {
	return u.databuffers
}

func (u *UdpServer) run() {
	log.Println("Listening on", u.conn.LocalAddr())
	const maxPacketSize = 6 << 10 // 64k
	for {
		data := <-u.databuffers
		msglen, _, err := u.conn.ReadFromUDP(data)
		if err != nil {
			log.Println("Udp error: ", err)
			break
		}
		u.packets <- data[:msglen]
	}
	log.Println("Udp server died.")
}

func (u *UdpServer) release(buffer []byte) {
	u.databuffers <- buffer
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
