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

package repository

import (
	"github.com/TimeWtr/generator/data_source"
	"github.com/TimeWtr/generator/domain"
	"golang.org/x/net/context"
)

type GeneratorRepository interface {
	// Insert 插入一条短码记录数据
	Insert(ctx context.Context, data domain.URLData, shardingKey int)
	// BatchInsert 批量插入同库同表记录数据
	BatchInsert(ctx context.Context, data []domain.URLData, shardingKey int)
}

type generatorRepositoryImpl struct {
	dataSource data_source.Factory
}
