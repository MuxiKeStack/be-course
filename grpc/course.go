package grpc

import (
	"context"
	coursev1 "github.com/MuxiKeStack/be-api/gen/proto/course/v1"
	"github.com/MuxiKeStack/be-course/domain"
	"github.com/MuxiKeStack/be-course/service"
	"github.com/ecodeclub/ekit/slice"
)

type CourseServiceServer struct {
	coursev1.UnimplementedCourseServiceServer
	svc service.CourseService
}

func (s *CourseServiceServer) List(ctx context.Context, request *coursev1.ListRequest) (*coursev1.ListResponse, error) {
	cs, err := s.svc.List(ctx, request.GetStudentId(), request.GetPassword(), request.GetYear(), request.GetTerm())
	return &coursev1.ListResponse{
		Courses: slice.Map(cs, func(idx int, src domain.FailOverCourse) *coursev1.Course {
			return convertToV(src)
		}),
	}, err
}

func convertToV(user domain.FailOverCourse) *coursev1.Course {
	return &coursev1.Course{
		CourseId: user.CourseId,
		Name:     user.Name,
		Teacher:  user.Teacher,
		Year:     user.Year,
		Term:     user.Term,
	}
}
