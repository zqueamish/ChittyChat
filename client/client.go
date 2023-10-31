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
	"unicode/utf8"

	cursor "atomicgo.dev/cursor"
	"github.com/inancgumus/screen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func joinChannel(ctx context.Context, client pb.ChatServiceClient) { //, Lamport int) {

	channel := pb.Channel{Name: *channelName, SendersName: *senderName}
	f := setLog(*senderName)
	defer f.Close()
	stream, err := client.JoinChannel(ctx, &channel)
	
	if err != nil {
		log.Fatalf("client.JoinChannel(ctx, &channel) throws: %v", err)
	}

	// Send the join message to the server
	sendMessage(ctx, client, "9cbf281b855e41b4ad9f97707efdd29d")
	// Channel that waits for the stream to close
	// So when <-waitc is called, it waits for an empty struct to be sent
	// This is to ensure that joinChannel doesn't exit before the stream's closed
	waitc := make(chan struct{})

	go func() {
		// The for loop is an infinite loop: Won't ever exit unless stream is closed
		for {
			// Receive message from stream and store it in incoming and possible error in err
			incoming, err := stream.Recv()

			// The stream closes when the client disconnects from the server, or vice versa
			// So if err == io.EOF, the stream is closed
			// If the stream is closed, close() is called on the waitc channel
			// Causeing <-waitc to exit, thus ending the joinChannel func
			if err == io.EOF {
				close(waitc)
				return
			}

			if err != nil {
				log.Fatalf("Failed to receive message from channel joining. \nError: %v", err)
			}

			incrLamport(incoming)

			messageFormat := "Received at " + formatClientMessage(incoming)

			if *senderName == incoming.GetSender() {
				if incoming.GetMessage() != fmt.Sprintf("Participant %v joined Chitty-Chat at Lamport time %v", incoming.GetSender(), incoming.GetTimestamp()-3) {
					clearPreviousConsoleLine()
				}
				log.Print(messageFormat)
				fmt.Print(messageFormat)
			} else {
				log.Print(messageFormat)
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
		fmt.Printf("Cannot send message - Error: %v", err)
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

	// Send message to server via stream
	stream.Send(&msg)

	// CloseAndRecv() closes the stream and returns the server's response, an ack
	ack, err := stream.CloseAndRecv()
	if err != nil {
		log.Printf("Cannot send message - Error: %v", err)
		fmt.Printf("Cannot send message - Error: %v", err)
	}

	log.Printf("Message  %v \n", ack)
	fmt.Printf("Message  %v \n", ack)
	// The prev. line is cleared, so that sent messages is not printed twice for the client
	clearPreviousConsoleLine()
}

// Function to increment the client's Lamport timestamp; used after receiving a message
func incrLamport(msg *pb.Message) {
	if msg.GetTimestamp() > Lamport {
		Lamport = msg.GetTimestamp() + 1
	} else {
		Lamport++
	}
	msg.Timestamp = Lamport
}

// Function from atomicgo.dev/cursor to clear the previous line in the console
func clearPreviousConsoleLine() {
	cursor.ClearLinesUp(1)
	cursor.StartOfLine()
}

// Function to format message to be printed to the client
func formatClientMessage(incoming *pb.Message) string {
	return fmt.Sprintf("Lamport time: %v\n[%v]: %v\n\n", incoming.GetTimestamp(), incoming.GetSender(), incoming.GetMessage())
}

func printWelcome() {
	fmt.Println("\n ━━━━━⊱⊱ ⋆  CHITTY CHAT ⋆ ⊰⊰━━━━━")
	fmt.Println("⋆｡˚ ☁︎ ˚｡ Welcome to " + *channelName)
	fmt.Println("⋆｡˚ ☁︎ ˚｡ Your username's " + *senderName)
	fmt.Println("⋆｡˚ ☁︎ ˚｡ To exit, press Ctrl + C\n\n")
}

var channelName = flag.String("channel", "Eepy", "Channel name for chatting")
var senderName = flag.String("username", "Anon", "Sender's name")
var tcpServer = flag.String("server", ":8080", "Tcp server")

var Lamport int32 = 0

func main() {
	screen.Clear()
	screen.MoveTopLeft()
	time.Sleep(time.Second / 60)

	flag.Parse()

	printWelcome()

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

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		message := scanner.Text()
		if !utf8.ValidString(message) {
			fmt.Printf("\n[Invalid characters.]\n[Please ensure your message is UTF-8 encoded.]\n\n")
			continue
		}
		if len(message) > 128 {
			fmt.Printf("\n[Brevity is the soul of wit.]\n[Please keep your message under 128 characters.]\n\n")
			continue
		}
		go sendMessage(ctx, client, message)
	}
}

    // sets the logger to use a log.txt file instead of the console
    func setLog(name string) *os.File {
        // Clears the log.txt file when a new server is started
		
		if _,err := os.Open(fmt.Sprintf("%s.txt",name)); err == nil {
        if err := os.Truncate(fmt.Sprintf("%s.txt",name), 0); err != nil {
            log.Printf("Failed to truncate: %v", err)
        }}

        // This connects to the log file/changes the output of the log information to the log.txt file.
        f, err := os.OpenFile(fmt.Sprintf("%s.txt",name), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
        if err != nil {
            log.Fatalf("error opening file: %v", err)
        }
        log.SetOutput(f)
        return f
    }