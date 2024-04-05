package client

import (
	"context"
	ccnuv1 "github.com/MuxiKeStack/be-api/gen/proto/ccnu/v1"
	"google.golang.org/grpc"
)

type RetryCCNUClient struct {
	ccnuv1.CCNUServiceClient
	retryCnt int
}

func (r *RetryCCNUClient) CourseList(ctx context.Context, in *ccnuv1.CourseListRequest, opts ...grpc.CallOption) (*ccnuv1.CourseListResponse, error) {
	var (
		res *ccnuv1.CourseListResponse
		err error
	)
	for i := 0; i < r.retryCnt; i++ {
		res, err = r.CCNUServiceClient.CourseList(ctx, in, opts...)
		if err == nil {
			return res, nil
		}
	}
	return nil, err
}
