#!/bin/bash

set -e

voskX8664URL="https://github.com/alphacep/vosk-api/releases/download/v0.3.45/vosk-linux-x86_64-0.3.45.zip"
voskARM64URL="https://github.com/alphacep/vosk-api/releases/download/v0.3.45/vosk-linux-aarch64-0.3.45.zip"

if [[ ! -d cmd ]]; then
	echo "Run this in the correct directory."
	exit 1
fi

function sttServicePrompt() {
    echo
    echo "Which speech-to-text service would you like to use?"
    echo "1: VOSK (fast, not as accurate)"
    echo "2: Whisper (slower, much more accurate, recommended for more powerful hardware)"
    echo
    read -p "Enter a number (2): " sttServiceNum
    if [[ ! -n ${sttServiceNum} ]]; then
        sttService="whisper"
    elif [[ ${sttServiceNum} == "1" ]]; then
        sttService="vosk"
    elif [[ ${sttServiceNum} == "2" ]]; then
        sttService="whisper"
    else
        echo
        echo "Choose a valid number, or just press enter to use the default number."
        sttServicePrompt
    fi
}

function voskModelPrompt() {
    echo
    echo "Which Vosk model would you like to use?"
    echo "1: en-US small"
    echo "2: en-US medium"
    echo "3: en-US large"
    echo
    read -p "Enter a number (2): " voskModelNum
    if [[ ! -n ${voskModelNum} ]]; then
        voskModel="vosk-model-small-en-us-0.15"
    elif [[ ${sttServiceNum} == "1" ]]; then
        voskModel="vosk-model-small-en-us-0.15"
    elif [[ ${sttServiceNum} == "2" ]]; then
        voskModel="vosk-model-en-us-0.22-lgraph"
    elif [[ ${sttServiceNum} == "3" ]]; then
        voskModel="vosk-model-en-us-0.22"
    else
        echo
        echo "Choose a valid number, or just press enter to use the default number."
        voskModelPrompt
    fi
}

function whichWhisperModel() {
    availableModels="tiny, small, medium, large-v3, large-v3-q5_0"
    echo
    echo "Which Whisper model would you like to use?"
    echo "Options: $availableModels"
    echo '(tiny is recommended, base is purposefully not here because it does not work well with short commands)'
    echo
    read -p "Enter preferred model: " whispermodel
    if [[ ! -n ${whispermodel} ]]; then
        echo
        echo "You must enter a key."
        whichWhisperModel
    fi
    if [[ ! ${availableModels} == *"${whispermodel}"* ]]; then
        echo
        echo "Invalid model."
        whichWhisperModel
    fi
}

function setupVosk() {
    if [[ ! -d vosklib ]]; then
        if [[ $(uname -m) == "aarch64" ]]; then
            wget "$voskARM64URL"
            unzip vosk-linux-aarch64-0.3.45.zip
            rm vosk-linux-aarch64-0.3.45.zip
            mv vosk-linux-aarch64-0.3.45 vosklib
        elif [[ $(uname -m) == "x86_64" ]]; then
            wget "$voskX8664URL"
            unzip vosk-linux-x86_64-0.3.45.zip
            rm vosk-linux-x86_64-0.3.45.zip
            mv vosk-linux-x86_64-0.3.45 vosklib
        else
            echo "Your architecture ($(uname -m)) is not supported."
            exit 1
        fi
    fi
    rm -rf vosk
    mkdir -p vosk/en-US
    cd vosk/en-US
    wget "https://alphacephei.com/vosk/models/${voskModel}.zip"
    unzip "${voskModel}.zip"
    mv ${voskModel} model
    rm ${voskModel}.zip
    cd ../..
}

function setupWhisper() {
    if [[ ! -d whisper.cpp ]]; then
        mkdir whisper.cpp
        cd whisper.cpp
        git clone https://github.com/ggerganov/whisper.cpp.git .
        git checkout 7fd6fa809749078aa00edf945e959c898f2bd1af
    else
        cd whisper.cpp
    fi
    ./models/download-ggml-model.sh $whispermodel
    mkdir -p ../whisper
    cp models/ggml-$whispermodel.bin ../whisper/ggml.bin
    rm -rf build_go
    cmake -B build_go \
    -DCMAKE_POSITION_INDEPENDENT_CODE=ON
    cmake --build build_go --config Release
    cd ..
}

function buildCavalier() {
    export CGO_ENABLED=1
    if [[ $sttService == "whisper" ]]; then
        export CGO_LDFLAGS="-L$(pwd)/whisper.cpp -L$(pwd)/whisper.cpp/build -L$(pwd)/whisper.cpp/build/src -L$(pwd)/whisper.cpp/build_go/ggml/src -L$(pwd)/whisper.cpp/build_go/src"
        export CGO_CFLAGS="-I$(pwd)/whisper.cpp -I$(pwd)/whisper.cpp/include -I$(pwd)/whisper.cpp/ggml/include"
    else
        export CGO_CFLAGS="-I$(pwd)/vosklib"
        export CGO_LDFLAGS="-L$(pwd)/vosklib -lvosk -ldl -lpthread"
    fi
    go build -ldflags "-w -s" -tags "nolibopusfile" -o cavalier cmd/${sttService}/main.go
    echo
    echo "cavalier has been compiled."
}

sttServicePrompt
if [[ ${sttService} == "vosk" ]]; then
    voskModelPrompt
    setupVosk
else
    whichWhisperModel
    setupWhisper
fi
buildCavalier
