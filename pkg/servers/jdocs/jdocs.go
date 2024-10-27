package jdocs

import (
	"context"
	"fmt"

	"github.com/digital-dream-labs/api/go/jdocspb"
)

type JdocServer struct {
	jdocspb.UnimplementedJdocsServer
}

func (s *JdocServer) WriteDoc(ctx context.Context, req *jdocspb.WriteDocReq) (*jdocspb.WriteDocResp, error) {
	fmt.Println("writedoc")
	fmt.Println(req)
	return nil, nil
}

func (s *JdocServer) ReadDoc(ctx context.Context, req *jdocspb.ReadDocsReq) (*jdocspb.ReadDocsResp, error) {
	fmt.Println("readdoc")
	fmt.Println(req)
	return nil, nil
}

func NewJdocsServer() *JdocServer {
	return &JdocServer{}
}
