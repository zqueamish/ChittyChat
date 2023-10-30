package main

import (
	"flag"
	"google.golang.org/grpc"
	"log"
	"net"
	proto "ChittyChat/grpc"
	"strconv"
)

// Struct that will be used to represent the Server.
type Server struct {
	proto.UnimplementedChatServiceServer // Necessary
	connections []proto.ChatService_SendMessagesServer // Connections to server
	name                             string
	port                             int
}

// Used to get the user-defined port for the server from the command line
var port = flag.Int("port", 0, "server port number")

func main() {
	// Get the port from the command line when the server is run
	flag.Parse()

	// Create a server struct
	server := &Server{
		name: "serverName",
		port: *port,
	}

	// Start the server
	go startServer(server)

	// Keep the server running until it is manually quit
	select {}
}

func startServer(server *Server) {

	// Create a new grpc server
	grpcServer := grpc.NewServer()

	// Make the server listen at the given port (convert int port to string)
	listener, err := net.Listen("tcp", "localhost:"+strconv.Itoa(server.port))

	if err != nil {
		log.Fatalf("Could not create the server %v", err)
	}
	log.Printf("Started server at port: %d\n", server.port)

	// Register the grpc server and serve its listener
	proto.RegisterChatServiceServer(grpcServer, server)
	serveError := grpcServer.Serve(listener)
	if serveError != nil {
		log.Fatalf("Could not serve listener")
	}
}

func (c *Server) SendMessages(msgStream proto.ChatService_SendMessagesServer) error {
    // Add the new connection to the list of active connections
    c.connections = append(c.connections, msgStream)

    // Wait for a message from the client
    for {
        msg, err := msgStream.Recv()
        if err != nil {
            log.Fatalf("Error when receiving message from client: %v", err)
        }
        // Print message on server
        log.Printf("%s says: %s", msg.Username, msg.Text)

        // Send the message to all connected clients
        for _, conn := range c.connections {
            if conn != msgStream {
                ack := &proto.Message{
                    Username: "server",
                    Text:     msg.Username + " says: " + msg.Text,
                }
                err := conn.Send(ack)
                if err != nil {
                    log.Fatalf("Error when sending message to client: %v", err)
                }
            }
        }
    }
}
