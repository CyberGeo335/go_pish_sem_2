package main

import (
	"context"
	"log"
	"time"

	"github.com/CyberGeo335/pz2-grpc/gen/studentpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func main() {
	conn, err := grpc.NewClient(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := studentpb.NewStudentServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pingResp, err := client.Ping(ctx, &studentpb.PingRequest{Message: "hello grpc"})
	if err != nil {
		log.Fatal("Ping error: ", err)
	}
	log.Println("Ping response:", pingResp.GetMessage())

	studentResp, err := client.GetStudentByID(ctx, &studentpb.GetStudentRequest{Id: 1})
	if err != nil {
		log.Fatal("GetStudentByID error: ", err)
	}
	printStudent("Student by id=1", studentResp.GetStudent())

	listResp, err := client.ListStudents(ctx, &emptypb.Empty{})
	if err != nil {
		log.Fatal("ListStudents error: ", err)
	}
	log.Printf("Students count: %d\n", len(listResp.GetStudents()))
	for _, st := range listResp.GetStudents() {
		printStudent("Student from list", st)
	}

	createdResp, err := client.CreateStudent(ctx, &studentpb.CreateStudentRequest{
		FullName:       "Смирнов Дмитрий Олегович",
		Group:          "ИВБО-04-25",
		Email:          "smirnov@example.com",
		Specialization: "Golang и микросервисы",
	})
	if err != nil {
		log.Fatal("CreateStudent error: ", err)
	}
	printStudent("Created student", createdResp.GetStudent())

	_, err = client.GetStudentByID(ctx, &studentpb.GetStudentRequest{Id: 999})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			log.Println("Expected error for id=999:", st.Code(), st.Message())
			return
		}
		log.Fatal("Unexpected error: ", err)
	}
}

func printStudent(prefix string, st *studentpb.Student) {
	log.Printf(
		"%s: id=%d, full_name=%s, group=%s, email=%s, specialization=%s\n",
		prefix,
		st.GetId(),
		st.GetFullName(),
		st.GetGroup(),
		st.GetEmail(),
		st.GetSpecialization(),
	)
}
