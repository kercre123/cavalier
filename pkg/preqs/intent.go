package processreqs

import (
	"fmt"
	"strings"

	sr "cavalier/pkg/speechrequest"
	ttr "cavalier/pkg/ttr"
	"cavalier/pkg/vars"
	"cavalier/pkg/vtt"
)

// This is here for compatibility with 1.6 and older software
func (s *Server) ProcessIntent(req *vtt.IntentRequest) (*vtt.IntentResponse, error) {
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
				fmt.Println("No intent was matched")
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
		// if vars.APIConfig.Knowledge.IntentGraph && vars.APIConfig.Knowledge.Enable {
		// 	fmt.Println("Making LLM request for device " + req.Device + "...")
		// 	_, err := ttr.StreamingKGSim(req, req.Device, transcribedText, false)
		// 	if err != nil {
		// 		fmt.Println("LLM error: " + err.Error())
		// 		logger.LogUI("LLM error: " + err.Error())
		// 		ttr.IntentPass(req, "intent_system_unmatched", transcribedText, map[string]string{"": ""}, false)
		// 		ttr.KGSim(req.Device, "There was an error getting a response from the L L M. Check the logs in the web interface.")
		// 	}
		// 	fmt.Println("Bot " + speechReq.Device + " request served.")
		// 	return nil, nil
		// }
		// fmt.Println("No intent was matched.")
		ttr.IntentPass(req, "intent_system_unmatched", transcribedText, map[string]string{"": ""}, false)
		return nil, nil
	}
	fmt.Println("Bot " + speechReq.Device + " request served.")
	return nil, nil
}
