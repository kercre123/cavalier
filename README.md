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

## What the accounts API does
- sends:

```
{
    "username": "redacted@gmail.com",
    "password": "redacted"
}
```

- receives: 
```
{
    "session": {
        "session_token": "redacted",
        "user_id": "redacted",
        "scope": "user",
        "time_created": "2024-10-26T00:26:57.174620948Z",
        "time_expires": "2025-10-26T00:26:57.174600148Z"
    },
    "user": {
        "user_id": "redacted",
        "drive_guest_id": "b80a7379-211a-4d7c-8440-01aa954635e1",
        "player_id": "b80a7379-211a-4d7c-8440-01aa954635e1",
        "created_by_app_name": null,
        "created_by_app_version": null,
        "created_by_app_platform": null,
        "dob": "1970-01-01",
        "email": "redacted@gmail.com",
        "family_name": null,
        "gender": null,
        "given_name": null,
        "username": "redacted@gmail.com",
        "email_is_verified": true,
        "email_failure_code": null,
        "email_lang": null,
        "password_is_complex": true,
        "status": "active",
        "time_created": "2024-10-24T18:42:56Z",
        "deactivation_reason": null,
        "purge_reason": null,
        "email_is_blocked": false,
        "no_autodelete": false,
        "is_email_account": true
    }
}
```

- if bad creds:
```
{
  "code": "server_failure",
  "message": "An unexpected server error occurred",
  "status": "error"
}
```

- components:
  1. sessions manager
       - creates session, expires when needed
  2. db comms
  3. ???
  4. profit (or not since this is an open-source project)

Why can't wire-pod run in the cloud?

- wire-pod relies on the IP addresses of incoming connections a lot and directly connects to robots via the SDK API. This cannot be done over the network without special configuration.
- Does wire-pod need to communicate with the bot over the bot's API? Well, something has to pull Jdocs from the robot in order to force the escape-pod token to be added to the store.
- **cavalier will not be compatible with production bots due to the workarounds required for escapepod firmware to consistently work with wire-pod.**
