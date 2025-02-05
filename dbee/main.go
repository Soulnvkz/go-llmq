package main

import (
	log "sol/proto/logger"

	dbconf "sol/dbee/db"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// import (
// 	"context"
// 	"fmt"
// 	"net"
// 	"sol/proto"

// 	"google.golang.org/grpc"
// )

// type server struct {
// 	proto.DbeeServer
// }

// func (s *server) SayHello(_ context.Context, in *proto.HelloRequest) (*proto.HelloReply, error) {
// 	fmt.Printf("Received: %v", in.GetName())
// 	return &proto.HelloReply{Message: "Hello " + in.GetName()}, nil
// }

// func main() {
// 	lis, err := net.Listen("tcp", ":5000")
// 	if err != nil {
// 		fmt.Printf("failed to listen: %v", err)
// 	}
// 	s := grpc.NewServer()

// 	proto.RegisterDbeeServer(s, &server{})
// 	fmt.Printf("server listening at %v", lis.Addr())
// 	if err := s.Serve(lis); err != nil {
// 		fmt.Printf("failed to serve: %v", err)
// 	}
// }

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Error().Fatal("failed to load .env file")
	}

	db_config := dbconf.NewDBConfig()
	db, err := db_config.Open()
	if err != nil {
		log.Error().Fatal("failed to open db connection")
	}

	log.Info().Println("connection succesfull...")

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Error().Fatal("failed to initilize postgres migrage")
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		log.Error().Fatal("failed to create migrate instance")
	}

	err = m.Up()
	if err != nil {
		log.Error().Fatal("failed to apply migrations")
	}

	log.Info().Println("migrations succesfull..")

	defer db.Close()
}
