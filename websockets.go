package main

import (
	"log"
    "strings"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var compData map[string]compEntry = make(map[string]compEntry);

type compEntry struct {
    Started bool
    ResetVotes map[string]int
}

func websocketTemplate(w http.ResponseWriter, r *http.Request, msgChannel chan string) {
	conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
		log.Println(err)
		return
	}
    for {
        messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
        msgChannel <- string(p)
        response := <- msgChannel
		if err := conn.WriteMessage(messageType, []byte(response)); err != nil {
			log.Println(err)
			return
		}
    }
}

func deployWS(c *gin.Context) {
    msgChannel := make(chan string)
    go websocketTemplate(c.Writer, c.Request, msgChannel)
	for {
		msg := <- msgChannel
    	_, err := vappDeployUser(msg, getUserName(c))
		if err != nil {
			msgChannel <- err.Error()
		} else {
			msgChannel <- "Deployed!"
		}
	}
}

func competitionEndpoint(c *gin.Context) {
    title := ""
    ev := Event{}
    msgChannel := make(chan string)
    go websocketTemplate(c.Writer, c.Request, msgChannel)
	for {
		msg := <- msgChannel
        // Check if image already deployed
        if title == "" {
            ev, _ = getEvent("title", msg)
            title = ev.Title
            if title == "" {
                msgChannel <- "Invalid title"
            } else {
                if _, ok := compData[title]; !ok {
                    compData[title] = compEntry{false, make(map[string]int)}
                    msgChannel <- "Deploy and Start"
                } else {
                    msgChannel <- "Ready"
                }
            }
        } else {
            // format: action:eventtitle:data
            if !validateCompWS(c, msg) {
                msgChannel <- "Oops, fraud detected"
            } else {
                cmdSlice := strings.Split(msg, ":")
                switch cmdSlice[0] {
                case "deploy":
            		id, err := vappDeployUser(ev.Vapp, getUserName(c))
            		if err != nil && err.Error() != "Already deployed" {
            			msgChannel <- err.Error()
            		} else {
            			msgChannel <- "Ready"
            		}
                    // power on
                    println("powering on", id)
                    // get ip
                    println("getting ips", id)
                case "reset":
                    // they reset lol
                case "unreset":
                    // they unreset lol
                }

            }
        }
	}
}

func validateCompWS(c *gin.Context, cmd string) bool {
    cmdSlice := strings.Split(cmd, ":")
    if len(cmdSlice) != 3 {
        return false
    }
    ev, err := getEvent("title", cmdSlice[1])
    if err != nil || getUserName(c) != ev.Field3 {
        // SQL error or not owner of lobby
        return false
    }
    return true
}
