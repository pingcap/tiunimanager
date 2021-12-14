/*******************************************************************************
 * @File: manager.go
 * @Description:
 * @Author: wangyaozheng@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/9
*******************************************************************************/

package management

import (
	"context"
	"github.com/pingcap-inc/tiem/library/common/resource-type"
)
type ResourceManager struct{}

func NewResourceManager() *ResourceManager {
	m := new(ResourceManager)
	return m
}

func (manager *ResourceManager)AllocResources(ctx context.Context, batchReq *resource.BatchAllocRequest) (results *resource.BatchAllocResponse, err error) {
	return nil, nil
}

func (manager *ResourceManager)RecycleResources(ctx context.Context, request *resource.RecycleRequest) (err error) {
	return nil
}