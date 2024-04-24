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

func (s *CourseServiceServer) Subscribed(ctx context.Context, request *coursev1.SubscribedRequest) (*coursev1.SubscribedResponse, error) {
	subscribed, err := s.svc.Subscribed(ctx, request.GetUid(), request.GetCourseId())
	return &coursev1.SubscribedResponse{
		Subscribed: subscribed,
	}, err
}

func NewCourseServiceServer(svc service.CourseService) *CourseServiceServer {
	return &CourseServiceServer{svc: svc}
}

func (s *CourseServiceServer) Register(server grpc.ServiceRegistrar) {
	coursev1.RegisterCourseServiceServer(server, s)
}

func (s *CourseServiceServer) SubscriptionList(ctx context.Context, request *coursev1.SubscriptionListRequest) (*coursev1.SubscriptionListResponse, error) {
	css, err := s.svc.SubscriptionList(ctx, request.GetStudentId(), request.GetPassword(),
		request.GetYear(), request.GetTerm(), request.GetUid()) // 传入了uid 说明这里肯定是调用带有容错或性能提升的List
	return &coursev1.SubscriptionListResponse{
		CourseSubscriptions: slice.Map(css, func(idx int, src domain.CourseSubscription) *coursev1.CourseSubscription {
			return convertToCourseSubscriptionV(src)
		}),
	}, err
}

func (s *CourseServiceServer) GetDetailById(ctx context.Context, request *coursev1.GetDetailByIdRequest) (*coursev1.GetDetailByIdResponse, error) {
	c, err := s.svc.GetDetailById(ctx, request.GetCourseId())
	return &coursev1.GetDetailByIdResponse{
		Course: convertToCourseV(c),
	}, err
}

func (s *CourseServiceServer) GetSubscriberUidsById(ctx context.Context,
	request *coursev1.GetSubscriberUidsByIdRequest) (*coursev1.GetSubscriberUidsByIdResponse, error) {
	uids, err := s.svc.GetSubscriberUidsByCourseId(ctx, request.GetCourseId(), request.GetCurUid(), request.GetLimit())
	return &coursev1.GetSubscriberUidsByIdResponse{
		InviteeUids: uids,
	}, err
}

func convertToCourseV(c domain.Course) *coursev1.Course {
	// 这里没有转换 year , term ，因为 year ， term 只在针对某个人查询的时候才有
	return &coursev1.Course{
		Id:         c.Id,
		CourseCode: c.CourseCode,
		Name:       c.Name,
		Teacher:    c.Teacher,
		School:     c.School,
		Property:   c.Property, // 发到外面就换成string，易于上游理解，内部是为了性能
		Credit:     c.Credit,
	}
}

func convertToCourseSubscriptionV(cs domain.CourseSubscription) *coursev1.CourseSubscription {
	return &coursev1.CourseSubscription{
		Course: &coursev1.Course{
			Id:         cs.Course.Id,
			CourseCode: cs.Course.CourseCode,
			Name:       cs.Course.Name,
			Teacher:    cs.Course.Teacher,
			School:     cs.Term,
			Property:   cs.Course.Property,
			Credit:     cs.Course.Credit,
		},
		Year: cs.Year,
		Term: cs.Term,
	}
}
