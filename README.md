# cavalier

This will maybe be a new Vector voice server implementation built from the ground-up to be able to run in the cloud rather than on a local device on the network.

This is mostly for my personal experiementation with firmwares below 1.2 which don't work with the latest vic-cloud code, so I can point them to a public cloud (as they require valid TLS certs).

Though, i'd like to be able to create something enterprise-ready which DDL could easily set up without much hassle if they wanted to.

Goals:

- Database storage (JSON storage an option as well)
- Accounts + working /v1/sessions
- Configs for different firmware versions
- TLS cert refresh
- VOSK 'n VAD

Notes:

- It's fine to use wire-pod components (speechrequest, the STT impls)

Why can't wire-pod run in the cloud?

- wire-pod relies on the IP addresses of incoming connections a lot and directly connects to robots via the SDK API. This cannot be done over the network without special configuration.
- Does wire-pod need to communicate with the bot over the bot's API? Well, something has to pull Jdocs from the robot in order to force the escape-pod token to be added to the store.
- **cavalier will not be compatible with production bots due to the workarounds required for escapepod firmware to consistently work with wire-pod.**
