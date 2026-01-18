package main

import (
	"cavalier/pkg/cavalier"
	stt "cavalier/pkg/whisper"
)

func main() {
	cavalier.InitCavalier(stt.Init, stt.STT, stt.Name)
}
