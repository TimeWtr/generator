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
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"math/rand"
	"testing"
	"time"
)

func TestNewDataFactory(t *testing.T) {
	dsn1 := "root:root@tcp(127.0.0.1:13306)/project?charset=utf8mb4&parseTime=True&loc=Local"
	db1, err := gorm.Open(mysql.Open(dsn1), &gorm.Config{})
	assert.Nil(t, err)

	dsn2 := "root:root@tcp(127.0.0.1:33061)/project?charset=utf8mb4&parseTime=True&loc=Local"
	db2, err := gorm.Open(mysql.Open(dsn2), &gorm.Config{})
	assert.Nil(t, err)

	dbs := []DataSource{
		{
			DB:         db1,
			TableCount: 10,
		},
		{
			DB:         db2,
			TableCount: 6,
		},
	}

	f := NewHashDataFactory(dbs, 20, "order_")
	for i := 0; i < 14; i++ {
		dst, err := f.GetDB(i)
		assert.Nil(t, err)
		t.Logf("dst message: %v", dst)
	}
}

func TestNewTimeDataSource(t *testing.T) {
	dsn1 := "root:root@tcp(127.0.0.1:13306)/project?charset=utf8mb4&parseTime=True&loc=Local"
	db1, err := gorm.Open(mysql.Open(dsn1), &gorm.Config{})
	assert.Nil(t, err)

	dsn2 := "root:root@tcp(127.0.0.1:33061)/project?charset=utf8mb4&parseTime=True&loc=Local"
	db2, err := gorm.Open(mysql.Open(dsn2), &gorm.Config{})
	assert.Nil(t, err)

	dbs := []TimeDataSource{
		{
			DS: DataSource{
				DB:         db1,
				TableCount: 10,
			},
			StartOffset: 0,
		},
		{
			DS: DataSource{
				DB:         db2,
				TableCount: 6,
			},
			StartOffset: 10,
		},
	}

	baseTime := time.Now().AddDate(-1, 0, 0)
	f := NewTimeDataSource(dbs, "order_", baseTime)
	for i := 0; i < 25; i++ {
		rd := rand.Intn(10)
		shardingKey := time.Now().AddDate(0, rd, 0)
		dst, er := f.GetDB(shardingKey)
		assert.Nil(t, er)
		t.Logf("dst message: %v", dst)
	}
}
