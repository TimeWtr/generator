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
	"github.com/TimeWtr/generator"
	"github.com/panjf2000/ants/v2"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"

	"github.com/TimeWtr/generator/repository/cache"
	"github.com/TimeWtr/generator/repository/dao"
	lmt "github.com/TimeWtr/local_message_table"

	"github.com/TimeWtr/Bitly/pkg/hs"

	intrv1 "github.com/TimeWtr/generator/api/proto/gen/intr.v1"
	"github.com/TimeWtr/generator/domain"
	"golang.org/x/net/context"
)

type URLServiceInter interface {
	// GenerateURL 生成单条URL
	GenerateURL(ctx context.Context, req *intrv1.URLRequest) (domain.URLResponse, error)
	// BatchGenerateURL 批量生成URL
	BatchGenerateURL(ctx context.Context, req *intrv1.BatchURLRequest) ([]domain.URLResponse, error)
}

const RetryCounts = 5

type Service struct {
	// ID获取的通道
	idCh <-chan int64
	// 数据库层操作
	d dao.ShortCodeInter
	// 缓存层操作
	cc cache.Cacher
	// 本地消息表
	lt lmt.MessagePusher
	// 全局的goroutine任务池
	pool *ants.Pool
}

func NewService(idCh <-chan int64, d dao.ShortCodeInter,
	cc cache.Cacher, lt lmt.MessagePusher, pool *ants.Pool) URLServiceInter {
	return &Service{
		idCh: idCh,
		d:    d,
		cc:   cc,
		lt:   lt,
		pool: pool,
	}
}

func (s *Service) GenerateURL(ctx context.Context, req *intrv1.URLRequest) (domain.URLResponse, error) {
	idHandler := NewIDHandler(s.idCh)
	hashHandler := NewHashHandler(hs.NewMurmur3())
	scHandler := NewShortCodeHandler(s.cc)
	dbHandler := NewDBHandler(s.lt, s.idCh)
	cmHandler := NewCompensateHandler(s.cc)
	idHandler.Next(hashHandler)
	hashHandler.Next(scHandler)
	scHandler.Next(dbHandler)
	dbHandler.Next(cmHandler)

	request := &Request{
		Biz:        req.GetBiz(),
		OriginURL:  req.GetMeta().GetOriginalUrl(),
		Creator:    req.GetCreator(),
		Comment:    req.GetMeta().GetComment(),
		Expiration: int(req.GetMeta().GetExpiration()),
		CustomCode: req.GetMeta().GetCustomCode(),
	}
	response := &Response{}
	err := idHandler.Process(ctx, request, response)
	if err != nil {
		return domain.URLResponse{}, err
	}

	return domain.URLResponse{
		ID:        response.ID,
		OriginURL: response.OriginURL,
		ShortCode: response.ShortCode,
		ExpireAt:  response.ExpireAt,
	}, nil
}

func (s *Service) BatchGenerateURL(ctx context.Context, req *intrv1.BatchURLRequest) ([]domain.URLResponse, error) {
	for _, r := range req.GetMeta() {
		request := &Request{
			Biz:        req.GetBiz(),
			OriginURL:  r.GetOriginalUrl(),
			Creator:    req.GetCreator(),
			Comment:    r.GetComment(),
			Expiration: int(r.GetExpiration()),
			CustomCode: r.GetCustomCode(),
		}

		s.pool.Submit(func() {
			idHandler := NewIDHandler(s.idCh)
			hashHandler := NewHashHandler(hs.NewMurmur3())
			scHandler := NewShortCodeHandler(s.cc)
			dbHandler := NewDBHandler(s.lt, s.idCh)
			cmHandler := NewCompensateHandler(s.cc)
			idHandler.Next(hashHandler)
			hashHandler.Next(scHandler)
			scHandler.Next(dbHandler)
			dbHandler.Next(cmHandler)

			response := &Response{}
			err := idHandler.Process(ctx, request, response)
			if err != nil {
				// 处理失败了
			}
		})
	}

	return nil, nil
}

func (s *Service) getID(ctx context.Context) (int64, error) {
	counter := 0
	for counter < RetryCounts {
		counter++

		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case id, ok := <-s.idCh:
			if !ok {
				return 0, errors.New("ID通道已关闭")
			}
			return id, nil
		}
	}

	return 0, errors.New("无可用ID")
}

// Handler 定义责任链处理短码生成的所有流程
// 雪花ID生成 -> URLHash计算，生成短码 -> 查询过滤器是否重复，重复则从短码池中获取一条预生成的可用短码
// 短码持久化到数据库 --> 如果持久化成功则向Kafka发送一条生成新短码的通知，跳转服务预加载到缓存
// --> 如果持久化失败，则进行补偿任务，将当前短码放回到短码池，并更新短码池短码计数
type Handler interface {
	Process(ctx context.Context, req *Request, resp *Response) error
	Next(Handler)
}

type BaseHandler struct {
	next Handler
}

func (b *BaseHandler) Next(h Handler) {
	b.next = h
}

type IDHandler struct {
	BaseHandler
	idCh <-chan int64
}

func NewIDHandler(idCh <-chan int64) Handler {
	return &IDHandler{
		idCh: idCh,
	}
}

func (i *IDHandler) Process(ctx context.Context, req *Request, resp *Response) error {
	if i.next == nil {
		return errors.New("哈希计算处理器不存在")
	}

	// 获取分布式ID
	var id int64
	counter := 0
	for counter < RetryCounts {
		counter++

		select {
		case <-ctx.Done():
			return ctx.Err()
		case newID, ok := <-i.idCh:
			if !ok {
				return errors.New("ID通道已关闭")
			}
			id = newID
			break
		}
	}

	if id == 0 {
		return errors.New("获取ID失败")
	}

	resp.ID = id
	return i.next.Process(ctx, req, resp)
}

type HashHandler struct {
	BaseHandler
	hs hs.Hasher
}

func NewHashHandler(hs hs.Hasher) Handler {
	return &HashHandler{
		hs: hs,
	}
}

// Process 调用Hash函数对原始URL进行hash计算，返回一个短码
func (h *HashHandler) Process(ctx context.Context, req *Request, resp *Response) error {
	if h.next == nil {
		return errors.New("短码验证处理器不存在")
	}

	shortCode, err := h.hs.ShortenURL(req.OriginURL)
	if err != nil {
		return err
	}

	resp.ShortCode = shortCode
	return h.next.Process(ctx, req, resp)
}

type ShortCodeHandler struct {
	BaseHandler
	cc cache.Cacher
	db dao.ShortCodeInter
}

func NewShortCodeHandler(cc cache.Cacher) Handler {
	return &ShortCodeHandler{
		cc: cc,
	}
}

// Process 将短码放入到过滤器中查询是否存在，如果"可能存在"，即假阳性，则需要到数据库中进行二次确认
// 如果不存在，则该短码为可用短码，反之则是重复短码，需要操作缓存从短码池中获取一条预生成的可用短码
func (s *ShortCodeHandler) Process(ctx context.Context, req *Request, resp *Response) error {
	if s.next == nil {
		return errors.New("数据库处理器不存在")
	}

	if resp.ShortCode != "" {
		res, err := s.cc.Exists(ctx, resp.ShortCode)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				res, err = s.cc.Exists(ctx, resp.ShortCode)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}

		// hash计算后的短码可直接使用
		if !res {
			return nil
		}
	}

	code, err := s.cc.GetShortCode(ctx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			code, err = s.cc.GetShortCode(ctx)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	resp.ShortCode = code

	return s.next.Process(ctx, req, resp)
}

// DBHandler 数据库处理器
type DBHandler struct {
	BaseHandler
	// 本地消息表机制
	lt lmt.MessagePusher
	// ID获取的通道
	idCh <-chan int64
}

func NewDBHandler(lt lmt.MessagePusher, idCh <-chan int64) Handler {
	return &DBHandler{
		lt:   lt,
		idCh: idCh,
	}
}

// Process 在该方法中需要传入一个可以执行执行的业务短码持久化的方法，并返回一个消息表Entity，推送和异步补偿机制都有
// 本地消息表来完成，实现持久化和消息稳定推送的一致性，短码具有唯一性，可以作为分库分表的分片键
func (d *DBHandler) Process(ctx context.Context, req *Request, resp *Response) error {
	if d.next == nil {
		return errors.New("未注册补偿任务处理器")
	}

	var err error
	defer func() {
		if err == nil {
			return
		}

		// 执行补偿任务
		er := d.next.Process(ctx, req, resp)
		if er != nil {
			//TODO 记录日志或发送一个监控告警
		}
	}()

	fn := func(ctx context.Context, tx *gorm.DB) (lmt.Messages, error) {
		id, er := d.getID(ctx)
		if er != nil {
			return lmt.Messages{}, er
		}

		now := time.Now().UnixMilli()
		expireAt := now + time.Duration(req.Expiration).Milliseconds()
		resp.ExpireAt = expireAt
		resp.ID = id

		er = tx.WithContext(ctx).Model(&dao.ShortCode{}).
			Create(&dao.ShortCode{
				OriginalURL: req.OriginURL,
				ShortCode:   resp.ShortCode,
				ExpireAt:    expireAt,
				Comment:     req.Comment,
				Creator:     req.Creator,
				CreateTime:  now,
				UpdateTime:  now,
			}).Error
		if err != nil {
			return lmt.Messages{}, err
		}

		id, er = d.getID(ctx)
		if er != nil {
			return lmt.Messages{}, er
		}

		var builder strings.Builder
		builder.WriteString("gen-")
		builder.WriteString(strconv.Itoa(int(id)))

		return lmt.Messages{
			ID:        id,
			Biz:       req.Biz,
			MessageID: builder.String(),
			Topic:     generator.Topic,
			Content:   builder.String(),
			Status:    lmt.MessageStatusNotSend.Int(),
		}, nil
	}

	err = d.lt.ExecTo(ctx, fn, resp.ShortCode)
	return err
}

func (d *DBHandler) getID(ctx context.Context) (int64, error) {
	counter := 0
	var id int64
	for counter < RetryCounts {
		counter++

		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case newID, ok := <-d.idCh:
			if !ok {
				return 0, errors.New("ID通道已关闭")
			}
			id = newID
			break
		}
	}
	if id == 0 {
		return 0, errors.New("获取ID失败")
	}

	return id, nil
}

// CompensateHandler 补偿处理器
type CompensateHandler struct {
	BaseHandler
	cc cache.Cacher
}

func NewCompensateHandler(cc cache.Cacher) Handler {
	return &CompensateHandler{
		cc: cc,
	}
}

func (c *CompensateHandler) Process(ctx context.Context, req *Request, resp *Response) error {
	return c.cc.InsertShortCode(ctx, resp.ShortCode)
}

type Response struct {
	ID        int64
	OriginURL string
	ShortCode string
	ExpireAt  int64
}

type Request struct {
	Biz        string
	Creator    string
	OriginURL  string
	CustomCode string
	Comment    string
	Expiration int
}
