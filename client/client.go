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
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func clearPreviousConsoleLine() {
	// \033[1A moves the cursor up one line
	// \033[2K clears the entire line
	fmt.Print("\033[1A\033[2K")
}

func formatClientMessage(incoming *pb.Message) string {
	return fmt.Sprintf("%v [%v]: %v\n", time.Now().Local().Format("15:04:05"), incoming.GetSender(), incoming.GetMessage())
}

func joinChannel(ctx context.Context, client pb.ChatServiceClient) {

	channel := pb.Channel{Name: *channelName, SendersName: *senderName}
	stream, err := client.JoinChannel(ctx, &channel)

	if err != nil {
		log.Fatalf("client.JoinChannel(ctx, &channel) throws: %v", err)
	}

	fmt.Println(" ")
	//fmt.Printf("Joined channel: %v\n", channel.Name)

	waitc := make(chan struct{})

	go func() {
		for {
			incoming, err := stream.Recv()

			if err == io.EOF {
				close(waitc)
				return
			}

			if err != nil {
				log.Fatalf("Failed to receive message from channel joining. \nError: %v", err)
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

func sendMessage(ctx context.Context, client pb.ChatServiceClient, message string) {
	stream, err := client.SendMessage(ctx)
	if err != nil {
		log.Printf("Cannot send message - Error: %v", err)
	}
	msg := pb.Message{
		Channel: &pb.Channel{
			Name:        *channelName,
			SendersName: *senderName},
		Message: message,
		Sender:  *senderName,
	}
	stream.Send(&msg)

	ack, err := stream.CloseAndRecv()
	fmt.Printf("Message  %v \n", ack) // Message Status: Sent if successful
	clearPreviousConsoleLine()
}

var channelName = flag.String("channel", "eepy chat", "Channel name for chatting")
var senderName = flag.String("sender", "default", "Sender's name")
var tcpServer = flag.String("server", ":8080", "Tcp server")

func main() {

	flag.Parse()

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

	go joinChannel(ctx, client)
	//var joinLeaveString = fmt.Sprintf("%v has joined the channel %v", *senderName, *channelName)
	//go sendMessage(ctx, client, joinLeaveString)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		go sendMessage(ctx, client, scanner.Text())
	}

}
