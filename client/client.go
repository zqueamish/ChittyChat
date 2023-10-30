package main

import (
	proto "ChittyChat/grpc"
	"bufio"
	"context"
	"flag"
	"log"
	"os"
	"strconv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	username  string
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

	// Prompt user to enter a username
	print("Enter a username: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	// Create a client
	client := &Client{
		username: scanner.Text(),
		id:         1,
		portNumber: *clientPort,
	}

	// Wait for the client (user) to Send messages
	go ClientMessageRequest(client)

	select {}
}

func ClientMessageRequest(client *Client) {
	// Connect to the server
	serverConnection, _ := connectToServer()

	// Wait for input in the client terminal
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()

		// Send message to server
		stream, err := serverConnection.SendMessages(context.Background())
		if err != nil {
			log.Print(err.Error())
		} else {
			message := &proto.Message{
				Username:  client.username,
				Text:      input,
				Timestamp: 0,
			}
			stream.Send(message)
			received, err := stream.Recv()
			if err != nil {
				log.Print(err.Error())
			} else {
				log.Printf(received.Text)
			}
		}
	}
}


func connectToServer() (proto.ChatServiceClient, error) {
	// Dial the server at the specified port.
	conn, err := grpc.Dial("localhost:"+strconv.Itoa(*serverPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to port %d", *serverPort)
	} else {
		log.Printf("Connected to the server at port %d\n", *serverPort)
	}
	return proto.NewChatServiceClient(conn), nil
}
