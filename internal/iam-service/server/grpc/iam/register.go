package iam

import (
	"context"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) RegisterByUsername(ctx context.Context, req *iam_service.RegisterByUsernameReq) (*emptypb.Empty, error) {
	if err := s.cli.RegisterByUsername(ctx, req.UserName, req.Password); err != nil {
		return nil, errStatus(errs.Code_IAMRegister, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) RegisterByEmail(ctx context.Context, req *iam_service.RegisterByEmailReq) (*emptypb.Empty, error) {
	if err := s.cli.RegisterByEmail(ctx, req.UserName, req.Email, req.Code); err != nil {
		return nil, errStatus(errs.Code_IAMRegister, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) RegisterSendEmailCode(ctx context.Context, req *iam_service.RegisterSendEmailCodeReq) (*emptypb.Empty, error) {
	if err := s.cli.RegisterSendEmailCode(ctx, req.UserName, req.Email); err != nil {
		return nil, errStatus(errs.Code_IAMRegister, err)
	}
	return &emptypb.Empty{}, nil
}
