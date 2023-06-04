/*
   Copyright 2022 Go-Flow Authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/basenana/go-flow/flow"
	"github.com/coreos/etcd/clientv3"
)

const (
	etcdFlowValueKeyPrefix = "storage.flow.value."
	etcdFlowValueKeyTpl    = etcdFlowValueKeyPrefix + "%s"
	etcdTaskKeyTpl         = "storage.flow.%s.task.%s"
)

type EtcdStorage struct {
	Client *clientv3.Client
}

func (e EtcdStorage) GetFlow(ctx context.Context, flowId string) (*flow.Flow, error) {
	var valueByte []byte
	if getValueResp, err := e.Client.Get(ctx, fmt.Sprintf(etcdFlowValueKeyTpl, flowId)); err != nil {
		return nil, err
	} else if len(getValueResp.Kvs) == 0 || len(getValueResp.Kvs) > 1 {
		return nil, fmt.Errorf("get no or more many flow value %s", flowId)
	} else {
		valueByte = getValueResp.Kvs[0].Value
	}

	result := &flow.Flow{}
	err := json.Unmarshal(valueByte, result)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal flow value, value: %v, err: %v", string(valueByte), err)
	}
	return result, nil
}

func (e EtcdStorage) GetFlows() ([]*flow.Flow, error) {
	getValueResp, err := e.Client.Get(context.TODO(), etcdFlowValueKeyPrefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	values := make([]*flow.Flow, len(getValueResp.Kvs))
	for i, v := range getValueResp.Kvs {
		f := &flow.Flow{}
		err = json.Unmarshal(v.Value, f)
		if err != nil {
			return nil, err
		}
		values[i] = f
	}
	return values, nil
}

func (e EtcdStorage) SaveFlow(ctx context.Context, flow *flow.Flow) error {
	valueByte, err := json.Marshal(flow)
	if err != nil {
		return err
	}

	_, err = e.Client.KV.Txn(ctx).Then(
		clientv3.OpPut(fmt.Sprintf(etcdFlowValueKeyTpl, flow.ID), string(valueByte)),
	).Commit()
	return err
}

func (e EtcdStorage) DeleteFlow(ctx context.Context, flowId string) error {
	_, err := e.Client.KV.Txn(ctx).Then(
		clientv3.OpDelete(fmt.Sprintf(etcdFlowValueKeyTpl, flowId)),
	).Commit()
	return err
}

func (e EtcdStorage) SaveTask(ctx context.Context, flowId string, task *flow.Task) error {
	taskByte, err := json.Marshal(task)
	if err != nil {
		return err
	}
	_, err = e.Client.KV.Txn(ctx).Then(
		clientv3.OpPut(fmt.Sprintf(etcdTaskKeyTpl, flowId, task.Name), string(taskByte)),
	).Commit()

	return err
}

func NewEtcdStorage(client *clientv3.Client) Interface {
	return &EtcdStorage{client}
}
