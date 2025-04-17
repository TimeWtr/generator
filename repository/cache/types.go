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

package cache

import "golang.org/x/net/context"

type Cacher interface {
	PoolCache
	BFCache
}

type PoolCache interface {
	// Count 短码池中的预生成短码数量
	Count(ctx context.Context) (int64, error)
	// GetShortCode 从短码池中获取一个预生成短码
	GetShortCode(ctx context.Context) (string, error)
	// InsertShortCode 向短码池中新增一条预生成短码
	InsertShortCode(ctx context.Context, code string) error
	// BatchInsertShortCodes 批量向短码池中增加多条预生成短码
	BatchInsertShortCodes(ctx context.Context, codes []string) error
}

type BFCache interface {
	// Reserve 创建一个新的过滤器
	Reserve(ctx context.Context, key string) error
	// Add 新增一条数据到过滤器
	Add(ctx context.Context, key string, data any) error
	// MAdd 批量新增数据到过滤器
	MAdd(ctx context.Context, key string, data []any) error
	// Exists 判断一条数据是否存在
	Exists(ctx context.Context, key string) (bool, error)
	// MExists 判断一批数据是否存在
	MExists(ctx context.Context, key string, data []any) (map[string]bool, error)
}
