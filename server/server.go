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

var timeFormat = time.Now().Local().Format("15:04:05") + " "

func (s *chatServiceServer) JoinChannel(ch *pb.Channel, msgStream pb.ChatService_JoinChannelServer) error {

	// Create a channel for the client
	clientChannel := make(chan *pb.Message)
	// Add the client-channel to the map
	s.channel[ch.Name] = append(s.channel[ch.Name], clientChannel)

	// Send a message to every client in the server that a new client has joined
	joinString := fmt.Sprintf("%v has joined the channel", ch.GetSendersName())
	fmt.Printf(timeFormat+"[%v]: "+joinString+"\n", ch.Name)
	msg := &pb.Message{Sender: ch.Name, Message: joinString, Channel: ch}
	go func() {
		streams := s.channel[msg.Channel.Name]
		for _, clientChan := range streams {
			clientChan <- msg
		}
	}()

	// doing this never closes the stream
	for {
		select {
		// if the client closes the stream / disconnects, the channel is closed
		case <-msgStream.Context().Done():
			leaveString := fmt.Sprintf("%v has left the channel", ch.GetSendersName())
			fmt.Printf(timeFormat+"[%v]: "+leaveString+"\n", ch.Name)

			// Remove the clientChannel from the slice of channels for this channel
			channels := s.channel[ch.Name]
			for i, channel := range channels {
				if channel == clientChannel {
					s.channel[ch.Name] = append(channels[:i], channels[i+1:]...)
					break
				}
			}

			// Send a message to every client in the channel that a client has left
			msg := &pb.Message{Sender: ch.Name, Message: leaveString, Channel: ch}
			go func() {
				streams := s.channel[msg.Channel.Name]
				for _, clientChan := range streams {
					clientChan <- msg
				}
			}()
			//}()

			// signals to server that the client has disconnected and the channel can be removed
			return nil
		// if a message is received from the client, send it to the server
		case msg := <-clientChannel:
			// sends the message to the client
			msgStream.Send(msg)
		}
	}
}

// Function to format message to be printed to the server
func formatMessage(msg *pb.Message) string {
	return fmt.Sprintf(timeFormat+"[%v]: %v\n", msg.Sender, msg.Message)
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
	fmt.Printf(formattedMessage)

	// Goroutine to send message to all clients in the channel
	go func() {
		streams := s.channel[msg.Channel.Name]
		for _, clientChan := range streams {
			clientChan <- msg
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
	fmt.Println("--- CHITTY CHAT ---")

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterChatServiceServer(grpcServer, &chatServiceServer{
		channel: make(map[string][]chan *pb.Message),
	})
	grpcServer.Serve(lis)
}
