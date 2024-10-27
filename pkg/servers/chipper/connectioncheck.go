package server

import (
	"context"
	"fmt"
	"strconv"
	"time"

	pb "github.com/digital-dream-labs/api/go/chipperpb"
)

const (
	connectionCheckTimeout = 15 * time.Second
	check                  = "check"
)

// StreamingConnectionCheck is used by the end device to make sure it can successfully communicate
func (s *Server) StreamingConnectionCheck(stream pb.ChipperGrpc_StreamingConnectionCheckServer) error {
	req, err := stream.Recv()
	fmt.Println("Incoming connection check from " + req.DeviceId)
	if err != nil {
		fmt.Println("Connection check unexpected error")
		fmt.Println(err)
		return err
	}

	ctx, cancel := context.WithTimeout(stream.Context(), connectionCheckTimeout)
	defer cancel()

	framesPerRequest := req.TotalAudioMs / req.AudioPerRequest

	var toSend pb.ConnectionCheckResponse

	// count frames, we already pulled the first one
	frames := uint32(1)
	toSend.FramesReceived = frames
receiveLoop:
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Connection check expiration. Frames Received: " + strconv.Itoa(int(frames)))
			toSend.Status = "Timeout"
			break receiveLoop
		default:
			req, suberr := stream.Recv()

			if suberr != nil || req == nil {
				err = suberr
				fmt.Println("Connection check unexpected error. Frames Received: " + strconv.Itoa(int(frames)))
				fmt.Println(err)

				toSend.Status = "Error"
				break receiveLoop
			}

			frames++
			toSend.FramesReceived = frames
			if frames >= framesPerRequest {
				fmt.Println("Connection check success")
				toSend.Status = "Success"
				break receiveLoop
			}
		}
	}
	senderr := stream.Send(&toSend)
	if senderr != nil {
		fmt.Println("Failed to send connection check response to client")
		fmt.Println(err)
		return senderr
	}
	return err

}
