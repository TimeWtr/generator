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

package hash

import (
	_ "embed"
	"errors"

	"github.com/TimeWtr/generator/repository/cache"

	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
)

//go:embed scripts/get_short_code.lua
var getShortCodeScript string

//go:embed scripts/set_short_code.lua
var setShortCodeScript string

//go:embed scripts/set_short_codes.lua
var setShortCodeArrayScript string

type CacheHash struct {
	client redis.Cmdable
}

func NewCacheHash(client redis.Cmdable) cache.Cacher {
	return &CacheHash{client: client}
}

func (c *CacheHash) Reserve(ctx context.Context, key string) error {
	//TODO implement me
	panic("implement me")
}

func (c *CacheHash) Add(ctx context.Context, key string, data any) error {
	//TODO implement me
	panic("implement me")
}

func (c *CacheHash) MAdd(ctx context.Context, key string, data []any) error {
	//TODO implement me
	panic("implement me")
}

func (c *CacheHash) Exists(ctx context.Context, key string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (c *CacheHash) MExists(ctx context.Context, key string, data []any) (map[string]bool, error) {
	//TODO implement me
	panic("implement me")
}

func (c *CacheHash) Count(ctx context.Context) (int64, error) {
	return c.client.Get(ctx, "shortCodeCount").Int64()
}

// GetShortCode 查询短码数量、获取一条可用的预生成短码、更新短码数量
func (c *CacheHash) GetShortCode(ctx context.Context) (string, error) {
	code := c.client.Eval(ctx, getShortCodeScript, []string{cache.PoolKey, cache.PoolLengthKey}).String()
	if code == "" {
		return "", errors.New("short code not found")
	}

	return code, nil
}

// InsertShortCode 新增一条新的预生成短码
func (c *CacheHash) InsertShortCode(ctx context.Context, code string) error {
	res, err := c.client.Eval(ctx, setShortCodeScript, []string{cache.PoolKey, cache.PoolLengthKey, cache.BFKey}, code).Int()
	if err != nil {
		return err
	}

	if res != 0 {
		return errors.New("failed to insert single short code to cache")
	}

	return nil
}

// BatchInsertShortCodes 预生成的短码批量写入缓存中
func (c *CacheHash) BatchInsertShortCodes(ctx context.Context, codes []string) error {
	res, err := c.client.Eval(ctx,
		setShortCodeArrayScript,
		[]string{cache.PoolKey, cache.PoolLengthKey, cache.BFKey},
		codes, len(codes)).Int()
	if err != nil {
		return err
	}

	if res != 0 {
		return errors.New("failed to insert batch short codes to cache")
	}

	return nil
}
