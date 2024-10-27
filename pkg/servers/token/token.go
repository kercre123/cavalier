package token

import (
	"context"
	"fmt"

	"github.com/digital-dream-labs/api/go/tokenpb"
)

type TokenServer struct {
	tokenpb.UnimplementedTokenServer
}

func (s *TokenServer) AssociatePrimaryUser(ctx context.Context, req *tokenpb.AssociatePrimaryUserRequest) (*tokenpb.AssociatePrimaryUserResponse, error) {
	fmt.Println("APU")
	fmt.Println(req)
	return nil, nil
}

func (s *TokenServer) AssociateSecondaryClient(ctx context.Context, req *tokenpb.AssociateSecondaryClientRequest) (*tokenpb.AssociateSecondaryClientResponse, error) {
	fmt.Println("ASU")
	fmt.Println(req)
	return nil, nil
}

func (s *TokenServer) RefreshToken(ctx context.Context, req *tokenpb.RefreshTokenRequest) (*tokenpb.RefreshTokenResponse, error) {
	fmt.Println("RT")
	fmt.Println(req)
	return nil, nil
}

func NewTokenServer() *TokenServer {
	return &TokenServer{}
}
