package main

import (
	"bufio"
	"context"
	"flag"
	"log"
	"os"
	proto "simpleGuide/grpc"
	"strconv"
	"time"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	id         int
	portNumber int
}

var (
	clientPort = flag.Int("cPort", 0, "client port number")
	serverPort = flag.Int("sPort", 0, "server port number (should match the port used for the server)")
)

func main() {
	// Parse the flags to get the port for the client
	flag.Parse()

	// Create a client
	client := &Client{
		id:         1,
		portNumber: *clientPort,
	}

	// Wait for the client (user) to ask for the time
	go waitForTimeRequest(client)

	for {

	}
}

func waitForTimeRequest(client *Client) {
	// Connect to the server
	serverConnection, _ := connectToServer()

	// Wait for input in the client terminal
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()
		log.Printf("Client asked for time with input: %s\n", input)

		// Ask the server for the time
		T1 := time.Now()
		timeReturnMessage, err := serverConnection.AskForTime(context.Background(), &proto.AskForTimeMessage{
			ClientId: int64(client.id),
		})
		if err != nil {
			log.Print(err.Error())
		} else {
			T4 := time.Now()
			T2 := time.Unix(0,timeReturnMessage.T2)
			T3 := time.Unix(0,timeReturnMessage.T3)
			roundtriptime := (T4.Sub(T1)) - (T3.Sub(T2))
			log.Println("Round trip time:", roundtriptime.Nanoseconds())
		}
	}
}

func connectToServer() (proto.TimeAskClient, error) {
	// Dial the server at the specified port.
	conn, err := grpc.Dial("localhost:"+strconv.Itoa(*serverPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to port %d", *serverPort)
	} else {
		log.Printf("Connected to the server at port %d\n", *serverPort)
	}
	return proto.NewTimeAskClient(conn), nil
}
