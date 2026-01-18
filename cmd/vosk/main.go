package main

import (
	"cavalier/pkg/cavalier"
	stt "cavalier/pkg/vosk"
)

func main() {
	cavalier.InitCavalier(stt.Init, stt.STT, stt.Name)
}
