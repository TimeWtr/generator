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

package service

import (
	"errors"
	"fmt"

	"github.com/TimeWtr/generator/repository/cache"
	"github.com/TimeWtr/generator/repository/dao"

	"github.com/TimeWtr/Bitly/pkg/hs"

	intrv1 "github.com/TimeWtr/generator/api/proto/gen/intr.v1"
	"github.com/TimeWtr/generator/domain"
	"golang.org/x/net/context"
)

type URLServiceInter interface {
	// GenerateURL 生成单条URL
	GenerateURL(ctx context.Context, req *intrv1.URLRequest) (domain.URLResponse, error)
	// BatchGenerateURL 批量生成URL
	BatchGenerateURL(ctx context.Context, req *intrv1.URLRequest) ([]domain.URLResponse, error)
}

const RetryCounts = 5

type Service struct {
	// ID获取的通道
	idCh <-chan int64
	// 数据库操作
	d dao.ShortCodeInter
	// 缓存
	cc cache.CInter
}

func NewService(idCh <-chan int64) URLServiceInter {
	return &Service{
		idCh: idCh,
	}
}

func (s *Service) GenerateURL(ctx context.Context, req *intrv1.URLRequest) (domain.URLResponse, error) {
	// 获取一个可用的短码
	code, err := s.cc.GetShortCode(ctx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			code, err = s.cc.GetShortCode(ctx)
			if err != nil {
				return domain.URLResponse{}, err
			}
		} else {
			return domain.URLResponse{}, err
		}
	}

	fmt.Println("code:", code)

	return domain.URLResponse{}, nil
}

func (s *Service) BatchGenerateURL(ctx context.Context, req *intrv1.URLRequest) ([]domain.URLResponse, error) {
	//TODO implement me
	panic("implement me")
}

// URLHandler 定义责任链+工厂接口
type URLHandler interface {
	Process(ctx context.Context, req *intrv1.URLRequest, resp *Response) error
	WithNext(next URLHandler) URLHandler
}

type IDHandler struct {
	idCh <-chan int64
	next URLHandler
}

func NewIDHandler(idCh <-chan int64) URLHandler {
	return &IDHandler{
		idCh: idCh,
	}
}

func (i *IDHandler) Process(ctx context.Context, req *intrv1.URLRequest, resp *Response) error {
	// 获取分布式ID
	var id int64
	counter := 0
	for counter < RetryCounts {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case id = <-i.idCh:
			break
		default:
			counter++
		}
	}
	if id == 0 {
		return errors.New("failed to get id")
	}

	resp.ID = id

	return i.next.Process(ctx, req, resp)
}

func (i *IDHandler) WithNext(next URLHandler) URLHandler {
	i.next = next
	return i
}

type HashHandler struct {
	hs   hs.Hasher
	next URLHandler
}

func NewHashHandler(hs hs.Hasher, next URLHandler) URLHandler {
	return &HashHandler{
		next: next,
	}
}

func (h *HashHandler) Process(ctx context.Context, req *intrv1.URLRequest, resp *Response) error {
	shortCode, err := h.hs.ShortenURL(req.GetMeta().GetOriginalUrl())
	if err != nil {
		return err
	}

	resp.ShortCode = shortCode
	return h.next.Process(ctx, req, resp)
}

func (h *HashHandler) WithNext(next URLHandler) URLHandler {
	h.next = next
	return h
}

type Response struct {
	ID        int64
	OriginURL string
	ShortCode string
	ExpireAt  int64
}
