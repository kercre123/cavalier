package server

import (
	"fmt"
	"time"

	"cavalier/pkg/vtt"

	pb "github.com/digital-dream-labs/api/go/chipperpb"
)

// StreamingIntent handles voice streams
func (s *Server) StreamingIntent(stream pb.ChipperGrpc_StreamingIntentServer) error {
	recvTime := time.Now()

	req, err := stream.Recv()
	if err != nil {
		fmt.Println("Intent error")
		fmt.Println(err)

		return err
	}

	if _, err = s.intent.ProcessIntent(
		&vtt.IntentRequest{
			Time:       recvTime,
			Stream:     stream,
			Device:     req.DeviceId,
			Session:    req.Session,
			LangString: req.LanguageCode.String(),
			FirstReq:   req,
			AudioCodec: req.AudioEncoding,
			// Mode:
		},
	); err != nil {
		fmt.Println("Intent error")
		fmt.Println(err)
		return err
	}

	return nil
}
