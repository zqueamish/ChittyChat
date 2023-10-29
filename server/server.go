package main

import (
	pb "ChittyChat/proto"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
)

// Struct contains a map of channels and a mutex to protect it.
// It stores the chan pointers contained in a single channel.
// A channel repr. our single client connected to the channel.
// Pointers are used to comm. between rpc calls JoinChannel and SendMessage.

type chatServiceServer struct {
	pb.UnimplementedChatServiceServer
	channel map[string][]chan *pb.Message
}

// JoinChannel function is called when a client joins a channel.
// When a client joins a channel, we add the channel to the map.

func (s *chatServiceServer) JoinChannel(ch *pb.Channel, msgStream pb.ChatService_JoinChannelServer) error {

	msgChannel := make(chan *pb.Message)
	s.channel[ch.Name] = append(s.channel[ch.Name], msgChannel)

	// doing this never closes the stream
	for {
		select {
		// if the client closes the stream / disconnects, the channel is closed
		case <-msgStream.Context().Done():
			// signals to server that the client has disconnected and the channel can be removed
			return nil
		// if a message is received from the client, send it to the server
		case msg := <-msgChannel:
			// prints the message to the server, note that the message is being printed x amount of times, where x is the number of clients connected to the channel
			//fmt.Printf("%v\n", formatMessage(msg))
			// sends the message to the client
			msgStream.Send(msg)
		}
	}
}

// Function to format message to be printed to the server
func formatMessage(msg *pb.Message) string {
	return fmt.Sprintf("[%v]: %v\n", msg.Sender, msg.Message)
}

// SendMessage function is called when a client sends a message.
// When a client sends a message, we send the message to the channel.

func (s *chatServiceServer) SendMessage(msgStream pb.ChatService_SendMessageServer) error {

	// Receive message from client
	msg, err := msgStream.Recv()

	// if there are no errors, return nil
	if err == io.EOF {
		return nil
	}

	// if there are errors, return the error
	if err != nil {
		return err
	}

	// Acknowledge message received to client
	ack := pb.MessageAck{Status: "Sent"}
	msgStream.SendAndClose(&ack)

	// Print message to server
	formattedMessage := formatMessage(msg)
	fmt.Printf(time.Now().Local().Format("15:04:05") + " " + formattedMessage)

	// Goroutine to send message to all clients in the channel
	go func() {
		streams := s.channel[msg.Channel.Name]
		for _, msgChan := range streams {
			msgChan <- msg
		}
	}()

	// Return nil to indicate success
	// This is required by the protobuf compiler
	// If we don't return nil, the protobuf compiler will throw an error
	// saying that the function doesn't return anything
	return nil

}

func main() {

	lis, err := net.Listen("tcp", ":8080")

	if err != nil {
		log.Fatalf("Failed to listen on port 8080: %v", err)
	}

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterChatServiceServer(grpcServer, &chatServiceServer{
		channel: make(map[string][]chan *pb.Message),
	})
	grpcServer.Serve(lis)
}
