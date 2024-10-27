package token

import (
	"context"
	"fmt"

	"github.com/digital-dream-labs/api/go/tokenpb"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type TokenServer struct {
	tokenpb.UnimplementedTokenServer
}

func getTransportCredentials(ctx context.Context) (*credentials.TLSInfo, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no peer info found in context")
	}
	fmt.Println(p)
	if p.AuthInfo != nil {
		if tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo); ok {
			if len(tlsInfo.State.PeerCertificates) == 0 {
				fmt.Println("no certs found :(")
				return nil, fmt.Errorf("no peer certificates found")
			}

			clientCert := tlsInfo.State.PeerCertificates[0]
			esn := clientCert.Subject.CommonName
			fmt.Println(esn)
			return &tlsInfo, nil
		}
	}

	return nil, fmt.Errorf("no transport credentials available")
}

func (s *TokenServer) AssociatePrimaryUser(ctx context.Context, req *tokenpb.AssociatePrimaryUserRequest) (*tokenpb.AssociatePrimaryUserResponse, error) {
	// map[:authority:[10.36.222.41:8081] anki-app-key:[oDoa0quieSeir6goowai7f] anki-user-session:[2BxvbWMxp3xGafZAdE4ZWP6] content-type:[application/grpc] user-agent:[Victor/v2.1.0.6-wire_os2.1.0.6-dev-202410220043 grpc-go/1.40.0]]
	fmt.Println("APU")
	getTransportCredentials(ctx)
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		fmt.Println("METADATA: ", md)
	} else {
		fmt.Println("no metadata, shame...")
	}
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
