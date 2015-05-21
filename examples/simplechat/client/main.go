package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jonas747/fnet"
	"github.com/jonas747/fnet/examples/simplechat"
	"github.com/jonas747/fnet/tcp"
	"os"
)

var addr = flag.String("addr", "home.jonas747.com:7449", "The address to listen on")

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

	hUserMsg, err := fnet.NewHandler(HandleMsg, int32(simplechat.Events_MESSAGE))
	panicErr(err)

	engine.AddHandler(hUserMsg)
	conn, err := tcp.Dial(*addr)
	panicErr(err)

	// Start all goroutines
	go engine.ListenChannels()
	go engine.HandleConn(conn)
	go listenErrors(engine)

	fmt.Println("Enter your name:")
	name := ""
	fmt.Scanln(&name)

	msg := &fnet.Message{
		EvtId: int32(simplechat.Events_USERJOIN),
		PB: &simplechat.User{
			Name: proto.String(name),
		},
	}

	err = conn.Send(msg)
	panicErr(err)

	for {
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		panicErr(err)
		line = line[:len(line)-1]
		msg := &fnet.Message{
			EvtId: int32(simplechat.Events_MESSAGE),
			PB: &simplechat.ChatMsg{
				Msg: proto.String(line),
			},
		}

		err = conn.Send(msg)
		panicErr(err)
	}
}

func HandleMsg(conn fnet.Connection, msg simplechat.ChatMsg) {
	fmt.Printf("[%s]: %s\n", msg.GetFrom(), msg.GetMsg())
}
