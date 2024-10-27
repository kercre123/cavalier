# var (
# 	SessionCertEnv = "SESSION_CERT_STORAGE"
# 	HoundKeyEnv    = "HOUND_KEY"
# 	HoundIDEnv     = "HOUND_ID"
# 	WeatherKey     = "WEATHER_KEY"
# )

if [[ ! -f source.sh ]]; then
    echo "source.sh must exist"
    exit 1
fi

source source.sh

CGO_ENABLED=1 CGO_CFLAGS="-I$(pwd)/vosklib" CGO_LDFLAGS="-L$(pwd)/vosklib" go build main.go

echo "built, running"

./main