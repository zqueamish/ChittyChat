package main

import (
	pb "ChittyChat/proto"
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

/*type client struct {
	pb.UnimplementedChatServiceServer

	Lamport int
}*/

func clearPreviousConsoleLine() {
	// \033[1A moves the cursor up one line
	// \033[2K clears the entire line
	fmt.Print("\033[1A\033[2K")
}

func formatClientMessage(incoming *pb.Message) string {
	return fmt.Sprintf("Lamport time: %v [%v]: %v\n", incoming.GetTimestamp(), incoming.GetSender(), incoming.GetMessage())
}

func joinChannel(ctx context.Context, client pb.ChatServiceClient) { //, Lamport int) {

	channel := pb.Channel{Name: *channelName, SendersName: *senderName}
	stream, err := client.JoinChannel(ctx, &channel)

	if err != nil {
		log.Fatalf("client.JoinChannel(ctx, &channel) throws: %v", err)
	}

	waitc := make(chan struct{})

	go func() {
		for {
			incoming, err := stream.Recv()

			// if it reaches the end of what it can receive, close the channel?
			if err == io.EOF {
				close(waitc)
				return
			}

			if err != nil {
				log.Fatalf("Failed to receive message from channel joining. \nError: %v", err)
			}

			// Lamport increment when message is received
			if incoming.GetTimestamp() > Lamport {
				incoming.Timestamp++
				Lamport = incoming.GetTimestamp()
			} else {
				Lamport++
				incoming.Timestamp = Lamport
			}

			messageFormat := formatClientMessage(incoming)

			if *senderName == incoming.GetSender() {
				clearPreviousConsoleLine()
				fmt.Print(messageFormat)
			} else {
				fmt.Print(messageFormat)
			}
		}
	}()

	<-waitc

}

func sendMessage(ctx context.Context, client pb.ChatServiceClient, message string) { //, Lamport int) {
	stream, err := client.SendMessage(ctx)
	if err != nil {
		log.Printf("Cannot send message - Error: %v", err)
	}
	// Increase Lamport timestamp before sending
	Lamport++

	// Create message
	msg := pb.Message{
		Channel: &pb.Channel{
			Name:        *channelName,
			SendersName: *senderName},
		Message: message,
		Sender:  *senderName,
		//Local Timestamp
		Timestamp: Lamport,
	}
	stream.Send(&msg)

	ack, err := stream.CloseAndRecv()
	if err != nil {
		log.Printf("Cannot send message - Error: %v", err)
	}

	fmt.Printf("Message  %v \n", ack) // Message Status: Sent if successful
	clearPreviousConsoleLine()
}

var channelName = flag.String("channel", "Eepy chat", "Channel name for chatting")
var senderName = flag.String("sender", "Anon", "Sender's name")
var tcpServer = flag.String("server", ":8080", "Tcp server")

var Lamport = 0

func main() {

	flag.Parse()
	//var Lamport = 0
	fmt.Println("--- CHITTY CHAT ---")

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.Dial(*tcpServer, opts...)
	if err != nil {
		log.Fatalf("Fail to dial: %v", err)
	}

	ctx := context.Background()
	client := pb.NewChatServiceClient(conn)

	defer conn.Close()

	go joinChannel(ctx, client) //, Lamport)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		go sendMessage(ctx, client, scanner.Text()) //, Lamport)
	}

}
