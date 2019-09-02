package main

import (
	"log"
	"time"

	pb "github.com/casbin/casbin-server/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	address = "192.168.50.249:50051"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewCasbinClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	
	a, err := c.NewAdapter(ctx, &pb.NewAdapterRequest{
		AdapterName: "TestAdapter", 
		DriverName: "mysql",
		ConnectString: "root:my-secret-pw@tcp(192.168.50.249:3306)/",
		TablePrefix: "zeus",
	})
	if err != nil {
		log.Fatalf("NewAdapter() error: %v", err)
	}
	log.Printf("Adapter handle: %s", a.Handler)

	r, err := c.NewEnforcer(ctx, &pb.NewEnforcerRequest{
		ModelText: "", 
		AdapterHandle: a.Handler,
		EnforcerName: "TestEnforcer",
	})
	if err != nil {
		log.Fatalf("NewEnforcer() error: %v", err)
	}
	log.Printf("Enforcer handle: %s", r.Handler)
}