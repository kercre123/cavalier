# cavalier

Pretty much fully functional cloud software for the Anki/DDL Vector robot.

Still somewhat experimental.

Can only be used with dev bots. I have an instance up at vicapi.pvic.xyz. My CFW ([victor](https://github.com/kercre123/victor)) is pointed to it. Vector Web Setup for it: https://v.pvic.xyz

## What is implemented?

- Accounts API (at port 8080)
- A sessions manager which expires tokens
- Full token and jdocs implementations (port 8081)
- SQLite3 storage for user credentials and bot jdocs
- Voice commands (chipper code copied from wire-pod) (also port 8081)
   - Weather, Houndify
- Rate limits

## Any differences between this and the DDL server software?

- The accounts endpoints are a bit different
  - /v1/sessions, /v1/create_user
- JWT tokens are not verified. This (I think) involves needing access to the per-bot cloud key database.

## TODO
- Email verification
- Reset password (function is there, just not in the API)
- More languages
- OpenAI?
- Crash dump upload (STS)

## how 2 run?

1. Put libvosk.so and vosk-api.h in a new directory called vosklib (can be downloaded from [here](https://github.com/alphacep/vosk-api/releases/tag/v0.3.45))
2. Create a source.sh file with the following:

```
export CERT=<path/to/fullchain.pem>
export KEY=<path/to/privkey.pem>
export WEATHER_KEY=<weatherapi.com key>
export HOUND_KEY=<houndify client key>
export HOUND_ID=<houndify client id>
```
3. Run start.sh. It will build the program and run it. If you are just running the program, make sure source.sh is sourced and LD_LIBRARY_PATH includes vosklib.
4. I use nginx as a proxy for the accounts API, and leave the rest not behind a proxy.
