package identification

import (
	"context"
	"github.com/google/uuid"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/message"
	"time"
)

// Provide provide a valid token string
func Provide (ctx context.Context, request message.ProvideTokenReq) (message.ProvideTokenResp, error) {
	// 提供token，简单地使用UUID
	tokenString := uuid.New().String()

	req := dbpb.DBSaveTokenRequest{
		Token: &dbpb.DBTokenDTO{
			TenantId:       request.Token.TenantID,
			AccountId:      request.Token.AccountID,
			AccountName:    request.Token.AccountName,
			ExpirationTime: request.Token.ExpirationTime.Unix(),
			TokenString:    tokenString,
		},
	}

	_, err := client.DBClient.SaveToken(ctx, &req)
	return message.ProvideTokenResp{TokenString: tokenString}, err
}

// Modify
func Modify (ctx context.Context, request message.ModifyTokenReq) error {
	req := dbpb.DBSaveTokenRequest{
		Token: &dbpb.DBTokenDTO{
			TenantId:       request.Token.TenantID,
			AccountId:      request.Token.AccountID,
			AccountName:    request.Token.AccountName,
			ExpirationTime: request.Token.ExpirationTime.Unix(),
			TokenString:    request.Token.TokenString,
		},
	}

	_, err := client.DBClient.SaveToken(ctx, &req)

	return err
}

// GetToken get token by tokenString
func GetToken(ctx context.Context, request message.GetTokenReq) (response message.GetTokenResp, err error) {
	req := dbpb.DBFindTokenRequest{
		TokenString: request.TokenString,
	}

	resp, err := client.DBClient.FindToken(ctx, &req)
	if err != nil {
		return
	}

	dto := resp.Token

	response.Token.TokenString = dto.TokenString
	response.Token.AccountID = dto.AccountId
	response.Token.AccountName = dto.AccountName
	response.Token.TenantID = dto.TenantId
	response.Token.ExpirationTime = time.Unix(dto.ExpirationTime, 0)

	return
}

