package account

import (
	"context"
	"fmt"
	dbCommon "github.com/pingcap-inc/tiem/models/common"
	"gorm.io/gorm"
	"strconv"
)

type AccountReadWrite struct {
	dbCommon.GormDB
}

func (g *AccountReadWrite) AddAccount(ctx context.Context, tenantId string, name string, salt string, finalHash string, status int8) (*Account, error) {
	if "" == tenantId || "" == name || "" == salt || "" == finalHash {
		return nil, fmt.Errorf("add account failed, has invalid parameter, tenantID: %s, name: %s, salt: %s, finalHash: %s, status: %d", tenantId, name, salt, finalHash, status)
	}
	result := &Account{
		Entity:    dbCommon.Entity{TenantId: tenantId, Status: strconv.Itoa(int(status))}, //todo: bug
		Name:      name,
		Salt:      salt,
		FinalHash: finalHash,
	}
	return result, g.DB(ctx).Create(result).Error
}

func (g *AccountReadWrite) FindAccountByName(ctx context.Context, name string) (*Account, error) {
	if "" == name {
		return nil, fmt.Errorf("find account failed, has invalid parameter, name: %s", name)
	}
	result := &Account{Name: name}
	return result, g.DB(ctx).Where(&Account{Name: name}).First(result).Error
}

func (g *AccountReadWrite) FindAccountById(ctx context.Context, id string) (*Account, error) {
	if "" == id {
		return nil, fmt.Errorf("find account failed, has invalid parameter, id: %s", id)
	}
	result := &Account{Entity: dbCommon.Entity{ID: id}}
	return result, g.DB(ctx).Where(&Account{Entity: dbCommon.Entity{ID: id}}).First(result).Error
}

func NewAccountReadWrite(db *gorm.DB) *AccountReadWrite {
	return &AccountReadWrite{
		dbCommon.WrapDB(db),
	}
}