# cavalier

Pretty much fully functional cloud software for the Anki/DDL Vector robot.

Still somewhat experimental.

Can only be used with unlocked bots (dev/oskr/unlock-prod). I have an instance up at vicapi.pvic.xyz. My CFW ([WireOS](https://github.com/os-vector/wire-os)) is pointed to it. Vector Web Setup for it: https://v.pvic.xyz

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
- JWT tokens are not verified. This (I think) requires access to the per-bot cloud key database.

## TODO
- Email verification
- Reset password (function is there, just not in the API)
- More languages
- OpenAI?
- Crash dump upload (STS)

## how 2 run?

- prerequisites: Go (1.21>), gcc, g++, make, cmake

1. cd into the cavalier directory, run `./setup.sh`
2. Create a source.sh file with the following:

```
export CERT=<path/to/fullchain.pem>
export KEY=<path/to/privkey.pem>
export WEATHER_KEY=<weatherapi.com key>
export HOUND_KEY=<houndify client key>
export HOUND_ID=<houndify client id>
```
And these optional settings if you desire:
```
export HOUND_IG=<true/false> (default is true)
export HOUND_LOCAL_TEXT_PROC=<true/false> (default is true)
```

3. Run cavalier.sh. It will run cavalier with the appropriate LD_LIBRARY_PATH, and with source.sh sourced.
4. I use nginx as a proxy for the accounts API, and leave the rest not behind a proxy.
