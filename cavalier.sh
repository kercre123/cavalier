# export CERT=<path/to/fullchain.pem>
# export KEY=<path/to/privkey.pem>
# export WEATHER_KEY=<weatherapi.com key>
# export HOUND_KEY=<houndify client key>
# export HOUND_ID=<houndify client id>

if [[ ! -f source.sh ]]; then
    echo "source.sh must exist in the same working directory as this script, see README for instructions"
    exit 1
fi

source source.sh

if [[ ! -f cavalier ]]; then
    echo "cavalier doesn't exist in the working directory. Either you haven't run setup.sh, or this script is running outside of the correct directory. Currently: $(pwd)."
    exit 1
fi

LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$(pwd)/vosklib:$(pwd)/whisper.cpp/build_go/src:$(pwd)/whisper.cpp/build_go/ggml/src ./cavalier