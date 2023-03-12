package main

import (
	"flag"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/lesismal/nbio/nbhttp/websocket"
)

var (
	onDataFrame      = flag.Bool("UseOnDataFrame", false, "Server will use OnDataFrame api instead of OnMessage")
	errBeforeUpgrade = flag.Bool("error-before-upgrade", false, "return an error on upgrade with body")
)

func newUpgrader() *websocket.Upgrader {
	u := websocket.NewUpgrader()

	if *onDataFrame {
		u.OnDataFrame(func(c *websocket.Conn, messageType websocket.MessageType, fin bool, data []byte) {
			// echo
			c.WriteFrame(messageType, true, fin, data)
			atomic.AddUint64(&qps, 1)
		})
	} else {
		u.OnMessage(func(c *websocket.Conn, messageType websocket.MessageType, data []byte) {
			// echo
			// fmt.Println(messageType, string(data))

			go func() {
				c.WriteMessage(messageType, data)
			}()
			atomic.AddUint64(&qps, 1)
		})
	}

	u.OnClose(func(c *websocket.Conn, err error) {
		fmt.Println("OnClose:", c.RemoteAddr().String(), err)
	})

	return u
}

func onWebsocket(w http.ResponseWriter, r *http.Request) {
	if *errBeforeUpgrade {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("returning an error"))
		return
	}

	upgrader := newUpgrader()
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}
	wsConn := conn.(*websocket.Conn)
	wsConn.SetReadDeadline(time.Time{})

	fmt.Println("OnOpen:", wsConn.RemoteAddr().String())
}
