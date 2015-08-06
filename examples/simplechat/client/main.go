package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jonas747/fnet"
	"github.com/jonas747/fnet/examples/simplechat"
	//"github.com/jonas747/fnet/tcp"
	"github.com/jonas747/fnet/ws"
	"os"
)

var addr = flag.String("addr", "ws://home.jonas747.com:7449", "The address to listen on")

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func listenErrors(engine *fnet.Engine) {
	for {
		err := <-engine.ErrChan
		fmt.Printf("fnet Error: ", err.Error())
	}
}

func main() {
	flag.Parse()
	fmt.Println("Running simplechat client!")
	engine := fnet.NewEngine()

	// stats
	//go simplechat.Monitor()

	hUserMsg, err := fnet.NewHandler(HandleMsg, int32(simplechat.Events_MESSAGE))
	panicErr(err)

	engine.AddHandler(hUserMsg)
	conn, err := ws.Dial(*addr, "", "http://localhost/")
	panicErr(err)

	// Start all goroutines
	go engine.ListenChannels()
	go engine.HandleConn(conn)
	go listenErrors(engine)

	fmt.Println("Enter your name:")
	name := ""
	fmt.Scanln(&name)

	msg := &simplechat.User{
		Name: proto.String(name),
	}

	raw, err := engine.CreateWireMessage(int32(simplechat.Events_USERJOIN), msg)
	panicErr(err)
	err = conn.Send(raw)
	panicErr(err)

	for {
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		panicErr(err)
		line = line[:len(line)-1]
		msg := &simplechat.ChatMsg{
			Msg: proto.String(line),
		}

		raw, err := engine.CreateWireMessage(int32(simplechat.Events_MESSAGE), msg)
		panicErr(err)
		err = conn.Send(raw)
		panicErr(err)
	}
}

func HandleMsg(conn fnet.Connection, msg simplechat.ChatMsg) {
	fmt.Printf("[%s]: %s\n", msg.GetFrom(), msg.GetMsg())
}
