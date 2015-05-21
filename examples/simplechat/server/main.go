package main

import (
	"flag"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jonas747/fnet"
	"github.com/jonas747/fnet/examples/simplechat"
	"github.com/jonas747/fnet/tcp"
)

var addr = flag.String("addr", ":7449", "The address to listen on")

var engine *fnet.Engine

func panicErr(errs ...error) {
	for _, v := range errs {
		if v != nil {
			panicErr(v)
		}
	}
}

func listenErrors() {
	for {
		err := <-engine.ErrChan
		fmt.Printf("fnet Error: ", err.Error())
	}
}

func main() {
	flag.Parse()
	fmt.Println("Running simplechat server!")

	// Initialize the handlers
	hUserJoin, err := fnet.NewHandler(HandleUserJoin, int32(simplechat.Events_USERJOIN))
	hUserLeave, err2 := fnet.NewHandler(HandleUserLeave, int32(simplechat.Events_USERLEAVE))
	hUserMsg, err3 := fnet.NewHandler(HandleSendMsg, int32(simplechat.Events_MESSAGE))
	panicErr(err, err2, err3)

	engine = fnet.NewEngine()
	engine.AddHandlers(hUserJoin, hUserLeave, hUserMsg)

	listener := &tcp.TCPListner{
		Engine: engine,
		Addr:   *addr,
	}

	// Start all goroutines
	go engine.ListenChannels()
	go engine.AddListener(listener)
	go listenErrors()

	engine.EmitConnOnClose = true

	// Code below to broadcast clients that has left
	for {
		c := <-engine.ConnCloseChan
		name, ok := c.GetSessionData().Get("name")
		if ok {
			nameStr := name.(string)
			chatMsg := fmt.Sprintf("\"%s\" Has left!", nameStr)
			msg := fnet.Message{
				PB: &simplechat.ChatMsg{
					From: proto.String("server"),
					Msg:  proto.String(chatMsg),
				},
				EvtId: int32(simplechat.Events_MESSAGE),
			}
			engine.Broadcast(msg)
		}
	}
}

func HandleUserJoin(conn fnet.Connection, user simplechat.User) {
	name := user.GetName()
	conn.GetSessionData().Set("name", name)
	msg := fnet.Message{
		EvtId: int32(simplechat.Events_MESSAGE),
		PB: &simplechat.ChatMsg{
			From: proto.String("server"),
			Msg:  proto.String("\"" + name + "\" Joined!"),
		},
	}
	engine.Broadcast(msg)
}

func HandleUserLeave(conn fnet.Connection, user simplechat.User) {
	fmt.Println("UserLeave!")
}

func HandleSendMsg(conn fnet.Connection, msg simplechat.ChatMsg) {
	name, _ := conn.GetSessionData().Get("name")
	nameStr := name.(string)
	reponse := fnet.Message{
		EvtId: int32(simplechat.Events_MESSAGE),
		PB: &simplechat.ChatMsg{
			From: proto.String(nameStr),
			Msg:  proto.String(msg.GetMsg()),
		},
	}
	engine.Broadcast(reponse)
}
