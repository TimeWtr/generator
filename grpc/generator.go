// Copyright 2025 TimeWtr
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package grpc

import (
	"errors"

	"github.com/TimeWtr/generator"
	intrv1 "github.com/TimeWtr/generator/api/proto/gen/intr.v1"
	"github.com/TimeWtr/generator/domain"
	"github.com/TimeWtr/generator/service"
	"golang.org/x/net/context"
)

type GeneratorServiceServer struct {
	intrv1.UnimplementedGeneratorServer
	srv service.URLServiceInter
}

func (g *GeneratorServiceServer) GenerateURL(ctx context.Context, req *intrv1.URLRequest) (*intrv1.URLResponse, error) {
	if req.GetBiz() == "" {
		return nil, errors.New("biz is required")
	}

	if req.GetMeta().GetOriginalUrl() == "" {
		return nil, errors.New("origin url is required")
	}

	if req.GetCreator() == "" {
		return nil, errors.New("creator is required")
	}

	expiration := req.GetMeta().GetExpiration()
	switch expiration {
	case generator.SevenDays, generator.FifteenDays, generator.ThirtyDays:
	default:
		return nil, errors.New("expiration is invalid")
	}

	res, err := g.srv.GenerateURL(ctx, req)
	if err != nil {
		return nil, err
	}

	return g.toDTO(res), nil
}

func (g *GeneratorServiceServer) BatchGenerateURL(ctx context.Context, req *intrv1.BatchURLRequest) (*intrv1.BatchURLResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (g *GeneratorServiceServer) UpdateURL(ctx context.Context, req *intrv1.URLRequest) (*intrv1.URLResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (g *GeneratorServiceServer) DeleteURL(ctx context.Context, req *intrv1.DelRequest) (*intrv1.DelResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (g *GeneratorServiceServer) mustEmbedUnimplementedGeneratorServer() {}

func (g *GeneratorServiceServer) toDTO(url domain.URLResponse) *intrv1.URLResponse {
	return &intrv1.URLResponse{
		StatusCode: 200,
		Message:    "generate success",
		Resp: &intrv1.URLResponseContent{
			OriginalUrl: url.OriginURL,
			ShortCode:   url.ShortCode,
			ExpireAt:    url.ExpireAt,
		},
	}
}
