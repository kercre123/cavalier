package jdocs

import (
	"cavalier/pkg/users"
	"cavalier/pkg/vars"
	"context"
	"errors"

	"github.com/digital-dream-labs/api/go/jdocspb"
)

type JdocServer struct {
	jdocspb.UnimplementedJdocsServer
}

func (s *JdocServer) WriteDoc(ctx context.Context, req *jdocspb.WriteDocReq) (*jdocspb.WriteDocResp, error) {
	if !users.IsRobotAssociatedWithAccount(req.Thing, req.UserId) {
		return nil, errors.New("not authorized")
	}
	vars.WriteJdoc(req.Thing, req.DocName, vars.AJdoc{
		DocVersion:     req.Doc.DocVersion,
		FmtVersion:     req.Doc.FmtVersion,
		ClientMetadata: req.Doc.ClientMetadata,
		JsonDoc:        req.Doc.JsonDoc,
	})
	return &jdocspb.WriteDocResp{
		Status:           jdocspb.WriteDocResp_ACCEPTED,
		LatestDocVersion: req.Doc.DocVersion,
	}, nil
}

func (s *JdocServer) ReadDocs(ctx context.Context, req *jdocspb.ReadDocsReq) (*jdocspb.ReadDocsResp, error) {
	if !users.IsRobotAssociatedWithAccount(req.Thing, req.UserId) {
		return nil, errors.New("not authorized")
	}
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
				Doc: &jdocspb.Jdoc{
					DocVersion:     0,
					FmtVersion:     0,
					ClientMetadata: "idontcare",
					JsonDoc:        "{}",
				},
			})
		}
	}
	return &resp, nil
}

func NewJdocsServer() *JdocServer {
	return &JdocServer{}
}
