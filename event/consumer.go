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

import (
	"github.com/ecodeclub/mq-api"
	"github.com/gotomicro/ego/core/elog"
)

type Consumer struct {
	// 处理器
	handler map[string]HandleFunc
	// kafka消费者
	consume mq.Consumer
	// 日志
	el *elog.Component
}

func NewSyncConsumer(q mq.MQ) (*Consumer, error) {
	const (
		topic   = "generator_events"
		groupId = "generator_group"
	)

	consumer, err := q.Consumer(topic, groupId)
	if err != nil {
		return nil, err
	}

	c := &Consumer{
		handler: make(map[string]HandleFunc),
		consume: consumer,
		el:      elog.DefaultLogger,
	}

	return c, nil
}
