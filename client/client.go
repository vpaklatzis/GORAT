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
	if conn, err = grpc.Dial(fmt.Sprintf("localhost:%d", 9090), opts...); err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client = grpcapi.NewAdminClient(conn)

	var cmd = new(grpcapi.Command)
	cmd.In = os.Args[1]
	ctx := context.Background()
	cmd, err = client.RunCommand(ctx, cmd)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(cmd.Out)
}
