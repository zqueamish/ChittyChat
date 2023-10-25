package main

import (
	"context"
	"flag"
	"google.golang.org/grpc"
	"log"
	"net"
	proto "simpleGuide/grpc"
	"strconv"
	"time"
)

// Struct that will be used to represent the Server.
type Server struct {
	proto.UnimplementedTimeAskServer // Necessary
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
	for {

	}
}

func startServer(server *Server) {

	// Create a new grpc server
	grpcServer := grpc.NewServer()

	// Make the server listen at the given port (convert int port to string)
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(server.port))

	if err != nil {
		log.Fatalf("Could not create the server %v", err)
	}
	log.Printf("Started server at port: %d\n", server.port)

	// Register the grpc server and serve its listener
	proto.RegisterTimeAskServer(grpcServer, server)
	serveError := grpcServer.Serve(listener)
	if serveError != nil {
		log.Fatalf("Could not serve listener")
	}
}

func (c *Server) AskForTime(ctx context.Context, in *proto.AskForTimeMessage) (*proto.TimeMessage, error) {
	var T2 int64 = time.Now().UnixNano()
	log.Printf("Client with ID %d asked for the time\n", in.ClientId)
	return &proto.TimeMessage{
		T2:       T2,
		T3: 	 time.Now().UnixNano(),
		ServerName: c.name,
	}, nil
}
