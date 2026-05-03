package student

import (
	"context"
	"strings"

	"github.com/CyberGeo335/pz2-grpc/gen/studentpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Service struct {
	studentpb.UnimplementedStudentServiceServer
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Ping(ctx context.Context, req *studentpb.PingRequest) (*studentpb.PingResponse, error) {
	msg := strings.TrimSpace(req.GetMessage())
	if msg == "" {
		msg = "ping"
	}

	return &studentpb.PingResponse{
		Message: "Server received: " + msg,
	}, nil
}

func (s *Service) GetStudentByID(ctx context.Context, req *studentpb.GetStudentRequest) (*studentpb.GetStudentResponse, error) {
	id := req.GetId()
	if id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid student id")
	}

	st, err := s.repo.GetByID(id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "student not found")
	}

	return &studentpb.GetStudentResponse{Student: st}, nil
}

func (s *Service) ListStudents(ctx context.Context, req *emptypb.Empty) (*studentpb.ListStudentsResponse, error) {
	return &studentpb.ListStudentsResponse{
		Students: s.repo.List(),
	}, nil
}

func (s *Service) CreateStudent(ctx context.Context, req *studentpb.CreateStudentRequest) (*studentpb.GetStudentResponse, error) {
	fullName := strings.TrimSpace(req.GetFullName())
	group := strings.TrimSpace(req.GetGroup())
	email := strings.TrimSpace(req.GetEmail())
	specialization := strings.TrimSpace(req.GetSpecialization())

	if fullName == "" {
		return nil, status.Error(codes.InvalidArgument, "full_name is required")
	}
	if group == "" {
		return nil, status.Error(codes.InvalidArgument, "group is required")
	}
	if email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	st := s.repo.Create(fullName, group, email, specialization)

	return &studentpb.GetStudentResponse{Student: st}, nil
}
