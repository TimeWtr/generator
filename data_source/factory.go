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

package data_source

import (
	"fmt"
	"github.com/TimeWtr/generator"
	"gorm.io/gorm"
	"time"
)

type Factory interface {
	// GetDB 根据分片获取链接
	GetDB(shardingKey any) (Dst, error)
}

// Dst 分片算法计算后落到的分片
type Dst struct {
	// 分库链接
	DB *gorm.DB
	// 分表
	Table string
}

type ShardType string

const (
	ShardTypeHash ShardType = "hash"
	ShardTypeTime ShardType = "time"
)

// DataSource 不同机器下库可以设置不同数量的分表，可以应对机器资源强度不一样的情况，
// 机器能力强的可以设置更多的分表，机器能力弱的可以设置少的分表。比如：
// db1: db1_order_1、db2_order_2、db3_order_3
// db2:	db1_order_1、db2_order_2
type DataSource struct {
	// 数据库链接
	DB *gorm.DB
	// 库中分表的数量
	TableCount int
}

// hashDataFactory 数据工厂模式，用于处理根据分片获取指定数据库链接
type hashDataFactory struct {
	// 库表映射关系
	dbs []DataSource
	// 总的库表数量
	totalTableCount int
	// 表前缀
	TablePrefix string
}

func NewHashDataFactory(dbs []DataSource, totalTableCount int, tablePrefix string) Factory {
	return &hashDataFactory{
		dbs:             dbs,
		totalTableCount: totalTableCount,
		TablePrefix:     tablePrefix,
	}
}

func (d *hashDataFactory) GetDB(shardingKey any) (Dst, error) {
	shardPos := shardingKey.(int) % d.totalTableCount
	currentPos := 0
	for _, ds := range d.dbs {
		// 当前的位置在当前分片区间内
		if currentPos < currentPos+ds.TableCount {
			tableIndex := shardPos - ds.TableCount
			return Dst{
				DB:    ds.DB,
				Table: fmt.Sprintf("%s%d", d.TablePrefix, tableIndex),
			}, nil
		}
		currentPos += ds.TableCount
	}

	return Dst{}, generator.ErrShardingFailed
}

// TimeDataSource 时间数据源的配置
type TimeDataSource struct {
	// 当前库的表数量对应的是负责处理几个月份的分表
	DS DataSource
	// 当前库中表的起始位置，标识是从第几个月开始
	StartOffset int
}

// timeDataFactory 基于时间来实现的分片算法实现
type timeDataFactory struct {
	// 数据库表信息
	dbs []TimeDataSource
	// 表名前缀
	TablePrefix string
	// 分库分表的基准时间，也就是开始时间，从哪个时间点来计算分库分表
	baseTime time.Time
}

func NewTimeDataSource(dbs []TimeDataSource, tablePrefix string, baseTime time.Time) Factory {
	return &timeDataFactory{
		dbs:         dbs,
		TablePrefix: tablePrefix,
		baseTime:    baseTime,
	}
}

func (t *timeDataFactory) GetDB(shardingKey any) (Dst, error) {
	st := shardingKey.(time.Time)
	months := int(st.Sub(t.baseTime).Hours() / 24 / 30)
	for _, d := range t.dbs {
		start := d.StartOffset
		end := start + d.DS.TableCount
		if months > start && months < end {
			return Dst{
				DB:    d.DS.DB,
				Table: fmt.Sprintf("%s%s", t.TablePrefix, st.Format("20060102")),
			}, nil
		}
	}

	return Dst{}, generator.ErrShardingFailed
}
