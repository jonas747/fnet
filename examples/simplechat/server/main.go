package main

import (
	"flag"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jonas747/fnet"
	"github.com/jonas747/fnet/examples/simplechat"
	"os"
	"runtime/pprof"

	//"github.com/jonas747/fnet/tcp"
	"github.com/jonas747/fnet/ws"
)

var addr = flag.String("addr", ":7449", "The address to listen on")
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var heapprofile = flag.String("heap", "", "write heap profile to file")

var engine *fnet.Engine

var signalChan = make(chan os.Signal, 2)

func panicErr(errs ...error) {
	for _, v := range errs {
		if v != nil {
			panicErr(v)
		}
	}
}

func listenErrors() {
	for {
		select {
		case err := <-engine.ErrChan:
			fmt.Printf("fnet Error: ", err.Error())
		case _ = <-signalChan:
			fmt.Println("Recived signal, ending...")
			return
		}
	}
}

func main() {
	flag.Parse()
	fmt.Println("Running simplechat server!")

	if *cpuprofile != "" {
		fmt.Println("Running with profiler enabled")
		f, err := os.Create(*cpuprofile)
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// Stats
	go simplechat.Monitor()

	engine = fnet.DefaultEngine()
	engine.OnConnClose = HandleConnectionClose
	engine.OnConnOpen = HandleConnectionOpen
	engine.Encoder = fnet.JsonEncoder

	// Initialize the handlers
	engine.AddHandler(fnet.NewHandlerSafe(HandleUserJoin, int32(simplechat.Events_USERJOIN)))
	engine.AddHandler(fnet.NewHandlerSafe(HandleUserLeave, int32(simplechat.Events_USERLEAVE)))
	engine.AddHandler(fnet.NewHandlerSafe(HandleSendMsg, int32(simplechat.Events_MESSAGE)))

	listener := &ws.WebsocketListener{
		Engine: engine,
		Addr:   *addr,
	}

	// Start all goroutines
	go engine.ListenChannels()
	go engine.AddListener(listener)
	go listenErrors()
	fmt.Scanln()
	if *heapprofile != "" {
		f, err := os.Create(*heapprofile)
		if err != nil {
			panic(err)
		}
		profile := pprof.Lookup("goroutine")
		profile.WriteTo(f, 1)
		f.Close()
	}
}

func HandleConnectionOpen(session fnet.Session) {
	fmt.Println("A connection opened!")
}

func HandleConnectionClose(session fnet.Session) {
	name := user.GetName()
	session.Data.Set("name", name)
	msg := &simplechat.ChatMsg{
		From: proto.String("server"),
		Msg:  proto.String("\"" + name + "\" Left! D:"),
	}
	err := engine.CreateAndBroadcast(int32(simplechat.Events_MESSAGE), msg)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	fmt.Println(name + " Left the chat! D:")
}

func HandleUserJoin(session fnet.Session, user simplechat.User) {
	name := user.GetName()
	session.Data.Set("name", name)
	msg := &simplechat.ChatMsg{
		From: proto.String("server"),
		Msg:  proto.String("\"" + name + "\" Joined!"),
	}

	fmt.Println(name + " Joined the chat!")

	err := engine.CreateAndBroadcast(int32(simplechat.Events_MESSAGE), msg)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
}

func HandleUserLeave(session fnet.Session, user simplechat.User) {
	fmt.Println("UserLeave!")
}

func HandleSendMsg(session fnet.Session, msg simplechat.ChatMsg) {
	name, _ := session.Data.GetString("name")
	response := &simplechat.ChatMsg{
		From: proto.String(name),
		Msg:  proto.String(msg.GetMsg()),
	}

	err := engine.CreateAndBroadcast(int32(simplechat.Events_MESSAGE), response)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
}
