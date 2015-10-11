package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jonas747/fnet"
	"github.com/jonas747/fnet/examples/simplechat"
	"time"
	//"github.com/jonas747/fnet/tcp"
	"github.com/jonas747/fnet/ws"
	"os"
)

var addr = flag.String("addr", "ws://home.jonas747.com:7449", "The address to listen on")
var bench = flag.Bool("bench", false, "Run a benchmark")
var benchInterval = flag.Float64("bint", 1, "Milliseconds between each message sent")

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
	engine := fnet.DefaultEngine()

	// stats
	//go simplechat.Monitor()

	engine.AddHandler(fnet.NewHandlerSafe(HandleMsg, int32(simplechat.Events_MESSAGE)))
	session, err := ws.Dial(*addr, "", "http://localhost/")
	panicErr(err)

	// Start all goroutines
	go engine.ListenChannels()
	go engine.HandleConn(session)
	go listenErrors(engine)

	// Set the name of the user from console inputs
	fmt.Println("Enter your name:")
	name := ""
	fmt.Scanln(&name)

	// Construct a mesasge and send it
	msg := &simplechat.User{
		Name: proto.String(name),
	}
	err = engine.CreateAndSend(session, int32(simplechat.Events_USERJOIN), msg)
	panicErr(err)

	if !*bench {
		for {
			// Read input for chat messages
			reader := bufio.NewReader(os.Stdin)
			line, err := reader.ReadString('\n')
			panicErr(err)
			line = line[:len(line)-1]

			// Construct and send the chat message
			msg := &simplechat.ChatMsg{
				Msg: proto.String(line),
			}
			err = engine.CreateAndSend(session, int32(simplechat.Events_MESSAGE), msg)
			panicErr(err)
		}
	} else {
		counter := 0
		for _ = range time.Tick(time.Duration(*benchInterval * float64(time.Millisecond))) {
			counter++
			msg := fmt.Sprintf("Test message #%d yo!", counter)

			fmt.Println("Sending: ", msg)
			// Construct and send the chat message
			pmsg := &simplechat.ChatMsg{
				Msg: proto.String(msg),
			}
			err = engine.CreateAndSend(session, int32(simplechat.Events_MESSAGE), pmsg)
			panicErr(err)
		}
	}
}

func HandleMsg(session fnet.Session, msg simplechat.ChatMsg) {
	fmt.Printf("[%s]: %s\n", msg.GetFrom(), msg.GetMsg())
}
