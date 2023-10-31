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
	Lamport int32 // Remote timestamp; keeps local time for newly joined users
}

// JoinChannel function is called when a client joins a server.
// When a client joins a server, we createn a channel for the client, and add the chan to the map.

func (s *chatServiceServer) JoinChannel(ch *pb.Channel, msgStream pb.ChatService_JoinChannelServer) error {

	// Create a channel for the client
	clientChannel := make(chan *pb.Message)

	// Add the client-channel to the map
	s.channel[ch.Name] = append(s.channel[ch.Name], clientChannel)

	// doing this never closes the stream
	for {
		select {
		// if the client closes the stream / disconnects, the channel is closed
		case <-msgStream.Context().Done():

			leaveString := fmt.Sprintf("%v has left the channel", ch.GetSendersName())

			// Remove the clientChannel from the slice of channels for this channel
			s.removeChannel(ch, clientChannel)

			// Simulate that the client has sent a farewellmessage to the server
			s.Lamport++

			// Send a message to every client in the channel that a client has left
			msg := &pb.Message{Sender: ch.Name, Message: leaveString, Channel: ch, Timestamp: s.Lamport}

			s.sendMsgToClients(msg)

			// closes the function by returning nil.
			return nil

		// if a message is received from the client, send it to the server
		case msg := <-clientChannel:

			s.incrLamport(msg)

			// sends the message to the server's SendMessage method
			msgStream.Send(msg)
		}
	}
}

// SendMessage function is called when a client sends a message.
// AKA this is where the serever receives a message from a client, that the other clients shall receive.
// When a client sends a message, we send the message to the channel.

func (s *chatServiceServer) SendMessage(msgStream pb.ChatService_SendMessageServer) error {

	// Receive message from client
	msg, err := msgStream.Recv()

	// if the stream is closed, return nil
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

	s.sendMsgToClients(msg)

	// Return nil to indicate success
	// This is required by the protobuf compiler
	// If we don't return nil, the protobuf compiler will throw an error
	// saying that the function doesn't return anything
	return nil
}

// Function to increase server's Lamport timestamp; used after receiving a message
func (s *chatServiceServer) incrLamport(msg *pb.Message) {
	if msg.GetTimestamp() > s.Lamport {
		s.Lamport = msg.GetTimestamp() + 1
	} else {
		s.Lamport++
	}
	msg.Timestamp = s.Lamport
}

// Function to remove the channel from the map after the client has left
func (s *chatServiceServer) removeChannel(ch *pb.Channel, currClientChannel chan *pb.Message) {
	channels := s.channel[ch.Name]
	for i, channel := range channels {
		if channel == currClientChannel {
			s.channel[ch.Name] = append(channels[:i], channels[i+1:]...)
			break
		}
	}
}

// Function to send message to all clients in the channel
func (s *chatServiceServer) sendMsgToClients(msg *pb.Message) {
	s.Lamport++
	msg.Timestamp = s.Lamport

	go func() {
		formattedMessage := formatMessage(msg)
		fmt.Printf("Received at " + formattedMessage)

		streams := s.channel[msg.Channel.Name]
		for _, clientChan := range streams {
			clientChan <- msg
		}
	}()
}

// Function to format message to be printed to the server
func formatMessage(msg *pb.Message) string {
	return fmt.Sprintf("Lamport time: %v [%v]: %v\n", msg.GetTimestamp(), msg.GetSender(), msg.GetMessage())
}

func main() {
	lis, err := net.Listen("tcp", ":8080")

	if err != nil {
		log.Fatalf("Failed to listen on port 8080: %v", err)
	}
	fmt.Println("--- CHITTY CHAT ---")

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterChatServiceServer(grpcServer, &chatServiceServer{
		channel: make(map[string][]chan *pb.Message),
		//Remote timestamp
		Lamport: 0,
	})

	fmt.Printf("Server started at Lamport time: %v\n", 0)

	grpcServer.Serve(lis)
}
