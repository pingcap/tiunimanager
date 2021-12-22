package identification

import (
	"context"
	"github.com/google/uuid"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/user/identification"
)

// Provide provide a valid token string
func Provide (ctx context.Context, request message.ProvideTokenReq) (message.ProvideTokenResp, error) {
	// 提供token，简单地使用UUID
	tokenString := uuid.New().String()
	req := &identification.Token{
		TenantId:       request.Token.TenantId,
		AccountId:      request.Token.AccountId,
		AccountName:    request.Token.AccountName,
		ExpirationTime: request.Token.ExpirationTime,
		TokenString:    tokenString,
	}

	_, err := models.GetTokenReaderWriter().AddToken(ctx, req.TokenString, req.AccountName, req.AccountId, req.TenantId, req.ExpirationTime)

	return message.ProvideTokenResp{TokenString: tokenString}, err
}

// Modify
func Modify (ctx context.Context, request message.ModifyTokenReq) error {

	req := &identification.Token{
		TenantId:       request.Token.TenantId,
		AccountId:      request.Token.AccountId,
		AccountName:    request.Token.AccountName,
		ExpirationTime: request.Token.ExpirationTime,
		TokenString:    request.Token.TokenString,
	}

	_, err := models.GetTokenReaderWriter().AddToken(ctx, req.TokenString, req.AccountName, req.AccountId, req.TenantId, req.ExpirationTime)


	return err
}

// GetToken get token by tokenString
func GetToken(ctx context.Context, request message.GetTokenReq) (response message.GetTokenResp, err error) {
	req := identification.Token{
		TokenString: request.TokenString,
	}

	resp, err := models.GetTokenReaderWriter().FindToken(ctx, req.TokenString)
	if err != nil {
		return
	}

	response.Token.TokenString = resp.TokenString
	response.Token.AccountId = resp.AccountId
	response.Token.AccountName = resp.AccountName
	response.Token.TenantId = resp.TenantId
	response.Token.ExpirationTime = resp.ExpirationTime

	return
}

