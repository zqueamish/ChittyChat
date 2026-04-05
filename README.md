# ChittyChat

eepy chat service.

To run and use Chitty-Chat:

1. Open two seperate terminals
2. In the first terminal, start the server by running the command `go run server/server.go`
3. In the second terminal, enter the server as a client by running the command `go run client/client.go -username <username>`
   - where `<username>` encapsulates your chosen username
5. More clients can be made:
   - Open another terminal, and run `go run client/client.go -username <newUsername>`
     - where `<newUsername>` is unique for ease of reading and testing
6. To let a client exit the chat, go to the client terminal and press `ctrl + c`
7. To close the server entirely, go to the server terminal and press `ctrl + c`
8. Log files for the users can be found in the client folder saved as `<username\>.txt`
9. Log files for the server can be found in the server folder saved as `Server.txt`

> (Sidenote: The servers port is set to 8080 in the code always, so if it's in use or blocked the server wont work unless the port is changed in the code)
