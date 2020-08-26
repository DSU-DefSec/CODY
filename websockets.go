package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var compData map[string]compEntry = make(map[string]compEntry)

type compEntry struct {
	Started    bool
	ResetVotes map[string]int
}

func websocketTemplate(w http.ResponseWriter, r *http.Request, msgChannel chan string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	go websocketReader(conn, msgChannel)
	go websocketWriter(conn, msgChannel)
}

func websocketReader(conn *websocket.Conn, msgChannel chan string) {
	for {
		_, p, err := conn.ReadMessage()
		fmt.Println("Reading", string(p))
		if err != nil {
			log.Println(err)
			return
		}
		msgChannel <- string(p)
	}
}

func websocketWriter(conn *websocket.Conn, msgChannel chan string) {
	for {
		response := <-msgChannel
		fmt.Println("Writing", response)
		if err := conn.WriteMessage(1, []byte(response)); err != nil {
			log.Println(err)
			return
		}
	}
}

func deployWS(c *gin.Context) {
	msgChannel := make(chan string)
	go websocketTemplate(c.Writer, c.Request, msgChannel)
	for {
		msg := <-msgChannel
		msgChannel <- "Working"
		id, err := vappDeploy(msg, getUser(c))
		if err != nil {
			msgChannel <- err.Error()
		} else if id != "" {
			msgChannel <- "vApp was deployed. Powering on..."
			ips, err := vappPowerAndIPs(id)
			if err != nil {
				msgChannel <- err.Error()
				return
			}
			msgChannel <- "ips"
			msgChannel <- ips
		} else {
			msgChannel <- "Deployed!"
		}
	}
}

func validateCompWS(c *gin.Context, cmd string) bool {
	cmdSlice := strings.Split(cmd, ":")
	if len(cmdSlice) != 3 {
		return false
	}
	/*
		ev, err := getEvent("title", cmdSlice[1])
		if err != nil || getUser(c) != ev.Field3 {
			// SQL error or not owner of lobby
			return false
		}
	*/
	return true
}
