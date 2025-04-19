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

package dao

import (
	"gorm.io/gorm"
	"time"

	"github.com/TimeWtr/shortlink-platform/generator/domain"
	"golang.org/x/net/context"
)

type ShortCodeInter interface {
	Insert(ctx context.Context, data domain.URLData) error
	Update(ctx context.Context, data domain.URLData) error
	GetURLByID(ctx context.Context, id int64) (ShortCode, error)
	GetURLByShortCode(ctx context.Context, shortCode string) (ShortCode, error)
	Delete(ctx context.Context, id int64) error
}

type ShortCodeDao struct {
	db *gorm.DB
}

func NewShortCodeDao(db *gorm.DB) ShortCodeInter {
	return &ShortCodeDao{db: db}
}

func (d *ShortCodeDao) Insert(ctx context.Context, data domain.URLData) error {
	now := time.Now().Unix()
	return d.db.WithContext(ctx).Create(&ShortCode{
		OriginalURL: data.OriginURL,
		ShortCode:   data.ShortCode,
		ExpireAt:    data.ExpireAt,
		Creator:     data.Creator,
		Comment:     data.Comment,
		CreateTime:  now,
		UpdateTime:  now,
	}).Error
}

func (d *ShortCodeDao) Update(ctx context.Context, data domain.URLData) error {
	return d.db.WithContext(ctx).
		Model(&ShortCode{}).
		Where("id = ?", data.ID).
		Updates(map[string]interface{}{
			"short_code":  data.ShortCode,
			"expire_at":   data.ExpireAt,
			"comment":     data.Comment,
			"creator":     data.Creator,
			"update_time": time.Now(),
		}).Error
}

func (d *ShortCodeDao) GetURLByID(ctx context.Context, id int64) (ShortCode, error) {
	var res ShortCode
	return res, d.db.WithContext(ctx).
		Model(&ShortCode{}).
		Where("id = ?", id).
		First(&res).Error
}

func (d *ShortCodeDao) GetURLByShortCode(ctx context.Context, shortCode string) (ShortCode, error) {
	var res ShortCode
	return res, d.db.WithContext(ctx).
		Model(&ShortCode{}).
		Where("short_code = ?", shortCode).
		First(&res).Error
}

func (d *ShortCodeDao) Delete(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).Delete(&ShortCode{}, id).Error
}

type ShortCode struct {
	ID          int64  `gorm:"column:id;type:bigint;autoIncrement;not null;primaryKey;comment:主键" json:"id"`
	OriginalURL string `gorm:"column:original_url;type:text;not null;comment:原始URL" json:"original_url"`
	ShortCode   string `gorm:"column:short_code;type:varchar(255);uniqueIndex:short_code_idx;not null;comment:短码" json:"short_code"`
	ExpireAt    int64  `gorm:"column:expire_at;type:bigint;not null;comment:过期时间" json:"expire_at"`
	Comment     string `gorm:"column:comment;type:text;not null;comment:备注" json:"comment"`
	Creator     string `gorm:"column:creator;type:varchar(255);not null;comment:创建者" json:"creator"`
	CreateTime  int64  `gorm:"column:create_time;type:bigint;not null;comment:创建时间" json:"create_time"`
	UpdateTime  int64  `gorm:"column:update_time;type:bigint;not null; comment:更新时间" json:"update_time"`
}
