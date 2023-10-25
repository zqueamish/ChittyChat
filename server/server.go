package main

import (
	pb "ChittyChat/proto"
	"fmt"
	"io"
	"log"
	"net"

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

	// create a channel for the client
	// this channel is used to send messages to the client
	// the channel is added to the map s.channel
	// so s.channel is a map of channel pointers
	// where a channel's pointer can be appended to the map

	msgChannel := make(chan *pb.Message)
	s.channel[ch.Name] = append(s.channel[ch.Name], msgChannel)

	// doing this never closes the stream
	for {
		select {
		// if the client closes the stream / disconnects, the channel is closed
		case <-msgStream.Context().Done():
			// signals to server that the client has disconnected and the channel can be removed
			return nil
		// if a message is received from the client, send it to the channel
		case msg := <-msgChannel:
			// printss the message to the server
			fmt.Printf("GO ROUTINE (got message): %v\n", msg)
			// sends the message to the client
			msgStream.Send(msg)
		}
	}
}

// SendMessage function is called when a client sends a message.
// When a client sends a message, we send the message to all clients in the channel.

func (s *chatServiceServer) SendMessage(msgStream pb.ChatService_SendMessageServer) error {

	// Receive message from client
	msg, err := msgStream.Recv()

	if err == io.EOF {
		return nil
	}

	if err != nil {
		return err
	}

	// Acknowledge message received to client
	ack := pb.MessageAck{Status: "SENT"}
	msgStream.SendAndClose(&ack)

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
