package processreqs

import (
	"fmt"
	"strings"

	"cavalier/pkg/vtt"

	sr "cavalier/pkg/speechrequest"
	ttr "cavalier/pkg/ttr"
	"cavalier/pkg/vars"
)

func (s *Server) ProcessIntentGraph(req *vtt.IntentGraphRequest) (*vtt.IntentGraphResponse, error) {
	var successMatched bool
	speechReq := sr.ReqToSpeechRequest(req)
	var transcribedText string
	if !isSti {
		var err error
		transcribedText, err = sttHandler(speechReq)
		if err != nil {
			ttr.IntentPass(req, "intent_system_noaudio", "voice processing error: "+err.Error(), map[string]string{"error": err.Error()}, true)
			return nil, nil
		}
		if strings.TrimSpace(transcribedText) == "" {
			ttr.IntentPass(req, "intent_system_noaudio", "", map[string]string{}, false)
			return nil, nil
		}
		successMatched = ttr.ProcessTextAll(req, transcribedText, vars.IntentList, speechReq.IsOpus)
	} else {
		intent, slots, err := stiHandler(speechReq)
		if err != nil {
			if err.Error() == "inference not understood" {
				fmt.Println("Bot " + speechReq.Device + " No intent was matched")
				ttr.IntentPass(req, "intent_system_unmatched", "voice processing error", map[string]string{"error": err.Error()}, true)
				return nil, nil
			}
			fmt.Println(err)
			ttr.IntentPass(req, "intent_system_noaudio", "voice processing error", map[string]string{"error": err.Error()}, true)
			return nil, nil
		}
		ttr.ParamCheckerSlotsEnUS(req, intent, slots, speechReq.IsOpus, speechReq.Device)
		return nil, nil
	}
	if !successMatched {
		// 	fmt.Println("No intent was matched.")
		// 	if vars.APIConfig.Knowledge.Enable && vars.APIConfig.Knowledge.Provider == "openai" && len([]rune(transcribedText)) >= 8 {
		// 		apiResponse := openaiRequest(transcribedText)
		// 		response := &pb.IntentGraphResponse{
		// 			Session:      req.Session,
		// 			DeviceId:     req.Device,
		// 			ResponseType: pb.IntentGraphMode_KNOWLEDGE_GRAPH,
		// 			SpokenText:   apiResponse,
		// 			QueryText:    transcribedText,
		// 			IsFinal:      true,
		// 		}
		// 		req.Stream.Send(response)
		// 		return nil, nil
		// 	}
		ttr.IntentPass(req, "intent_system_unmatched", transcribedText, map[string]string{"": ""}, false)
		return nil, nil
	}
	fmt.Println("Bot " + speechReq.Device + " request served.")
	return nil, nil
}
