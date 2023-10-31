# ChittyChat

eepy chat service.

To run and use Chitty-Chat:

1. Open two seperate terminals
2. In the first terminal, start the server by running the command 'go run server/server.go'
3. In the second terminal, enter the sever as a client by running the command 'go run client/client.go -username \<username\>', where \<username\> encapsulates your chosen username.
4. More clients can be made. open another terminal, and run the client.go file with a new username, to send messages from another client to the server.
5. To let a client exit the chat, go to the client terminal, and simply press ctrl + c.
6. To close the server entirely, go to the server terminal, and press ctrl + c.
7. Log files for the users can be found in the client folder saved as \<username\>.txt
8. Log files for the server can be found in the server folder saved as Server.txt

(sidenote The servers port is set to 8080 in the code always so if its in use or blocked the server wont work unless the port is changed in the code)
