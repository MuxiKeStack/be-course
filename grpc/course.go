package grpc

import (
	"context"
	coursev1 "github.com/MuxiKeStack/be-api/gen/proto/course/v1"
	"github.com/MuxiKeStack/be-course/domain"
	"github.com/MuxiKeStack/be-course/service"
	"github.com/ecodeclub/ekit/slice"
	"google.golang.org/grpc"
)

type CourseServiceServer struct {
	coursev1.UnimplementedCourseServiceServer
	svc service.CourseService
}

func NewCourseServiceServer(svc service.CourseService) *CourseServiceServer {
	return &CourseServiceServer{svc: svc}
}

func (s *CourseServiceServer) Register(server grpc.ServiceRegistrar) {
	coursev1.RegisterCourseServiceServer(server, s)
}

func (s *CourseServiceServer) List(ctx context.Context, request *coursev1.ListRequest) (*coursev1.ListResponse, error) {
	cs, err := s.svc.List(ctx, request.GetStudentId(), request.GetPassword(),
		request.GetYear(), request.GetTerm(), request.GetUid()) // 传入了uid 说明这里肯定是调用带有容错或性能提升的List
	return &coursev1.ListResponse{
		Courses: slice.Map(cs, func(idx int, src domain.Course) *coursev1.Course {
			return convertToCourseV(src)
		}),
	}, err
}

func (s *CourseServiceServer) GetDetailById(ctx context.Context, request *coursev1.GetDetailByIdRequest) (*coursev1.GetDetailByIdResponse, error) {
	c, err := s.svc.GetDetailById(ctx, request.GetCourseId())
	return &coursev1.GetDetailByIdResponse{
		Course: convertToCourseV(c),
	}, err
}

func (s *CourseServiceServer) GetGradesById(ctx context.Context, request *coursev1.GetGradesByIdRequest) (*coursev1.GetGradesByIdResponse, error) {
	grades, err := s.svc.GetGradesById(ctx, request.GetCourseId())
	return &coursev1.GetGradesByIdResponse{Grades: slice.Map(grades, func(idx int, src domain.Grade) *coursev1.Grade {
		return convertToGradeV(src)
	})}, err
}

func convertToGradeV(grade domain.Grade) *coursev1.Grade {
	return &coursev1.Grade{
		Regular: grade.Regular,
		Final:   grade.Final,
		Total:   grade.Total,
		Year:    grade.Year,
		Term:    grade.Term,
	}
}

func convertToCourseV(c domain.Course) *coursev1.Course {
	return &coursev1.Course{
		Id:         c.Id,
		CourseCode: c.CourseCode,
		Name:       c.Name,
		Teacher:    c.Teacher,
		School:     c.School,
		Property:   c.Property.String(), // 发到外面就换成string，易于上游理解，内部是为了性能
		Credit:     c.Credit,
	}
}
