package testext

import (
	"context"
	"fmt"
	"strings"

	"github.com/bridgekit-io/frodo/fail"
)

type OtherServiceHandler struct {
	Sequence      *Sequence
	SampleService SampleService
}

func (svc OtherServiceHandler) SpaceOut(_ context.Context, req *OtherRequest) (*OtherResponse, error) {
	svc.Sequence.Append("SpaceOut:" + req.Text)
	runes := strings.Split(req.Text, "")
	return &OtherResponse{Text: strings.Join(runes, " ")}, nil
}

func (svc OtherServiceHandler) RPCExample(ctx context.Context, req *OtherRequest) (*OtherResponse, error) {
	svc.Sequence.Append("RPCExample:" + req.Text)
	res, err := svc.SampleService.TriggerUpperCase(ctx, &SampleRequest{Text: req.Text})
	if err != nil {
		return nil, fmt.Errorf("wtf: %w", err)
	}
	return &OtherResponse{Text: res.Text}, nil
}

func (svc OtherServiceHandler) ListenWell(_ context.Context, req *OtherRequest) (*OtherResponse, error) {
	svc.Sequence.Append("ListenWell:" + req.Text)
	return &OtherResponse{Text: "ListenWell:" + req.Text}, nil
}

func (svc OtherServiceHandler) ChainOne(_ context.Context, req *OtherRequest) (*OtherResponse, error) {
	svc.Sequence.Append("ChainOne:" + req.Text)
	return &OtherResponse{Text: "ChainOne:" + req.Text}, nil
}

func (svc OtherServiceHandler) ChainTwo(_ context.Context, req *OtherRequest) (*OtherResponse, error) {
	svc.Sequence.Append("ChainTwo:" + req.Text)
	return &OtherResponse{Text: "ChainTwo:" + req.Text}, nil
}

func (svc OtherServiceHandler) ChainThree(_ context.Context, req *OtherRequest) (*OtherResponse, error) {
	svc.Sequence.Append("ChainThree:" + req.Text)
	return &OtherResponse{Text: "ChainThree:" + req.Text}, nil
}

func (svc OtherServiceHandler) ChainFour(_ context.Context, req *OtherRequest) (*OtherResponse, error) {
	svc.Sequence.Append("ChainFour:" + req.Text)
	return &OtherResponse{Text: "ChainFour:" + req.Text}, nil
}

func (svc OtherServiceHandler) ChainFail(_ context.Context, req *OtherRequest) (*OtherResponse, error) {
	svc.Sequence.Append("ChainFail:" + req.Text)
	return nil, fail.Unexpected("no soup for you")
}

func (svc OtherServiceHandler) ChainFailAfter(_ context.Context, req *OtherRequest) (*OtherResponse, error) {
	svc.Sequence.Append("ChainFailAfter:" + req.Text)
	return &OtherResponse{Text: "ChainFailAfter:" + req.Text}, nil
}
