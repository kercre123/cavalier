package jdocs

import (
	"cavalier/pkg/vars"
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
	var resp jdocspb.ReadDocsResp
	for _, item := range req.Items {
		ajdoc, err := vars.ReadJdoc(req.Thing, item.DocName)
		if err == nil {
			jdoc := jdocspb.Jdoc{
				DocVersion:     ajdoc.DocVersion,
				FmtVersion:     ajdoc.FmtVersion,
				ClientMetadata: ajdoc.ClientMetadata,
				JsonDoc:        ajdoc.JsonDoc,
			}
			resp.Items = append(resp.Items, &jdocspb.ReadDocsResp_Item{
				Status: jdocspb.ReadDocsResp_CHANGED,
				Doc:    &jdoc,
			})
		} else {
			resp.Items = append(resp.Items, &jdocspb.ReadDocsResp_Item{
				Status: jdocspb.ReadDocsResp_NOT_FOUND,
			})
		}
	}
	fmt.Println("readdoc")
	fmt.Println(req)
	return nil, nil
}

func NewJdocsServer() *JdocServer {
	return &JdocServer{}
}
