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

package event

import "golang.org/x/net/context"

type Event struct {
	// 所属的任务ID
	TaskID string `json:"task_id,omitempty"`
	// 原始的URL
	OriginalURL string `json:"original_url,omitempty"`
	// 回调地址
	CallbackURL string `json:"callback_url,omitempty"`
}

type HandleFunc func(ctx context.Context, evt *Event) error
