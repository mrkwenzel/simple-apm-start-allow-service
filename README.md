# Simple APM "Start -> Allow" Service
This simple implementation should help debugging and testing of applications
relying on working with an APM. It should help implementing the right behavior
working with the `MRHSession` cookie and working with the ambigous `302`
redirect.

Currently works only with `http` but can be extended to `https`.

## How to run
You can set up following vars:

`SIMPLE_APM_PORT`: Defaults to `8080` but can be set up to any port.

`SIMPLE_APM_PROXIED_SERVICE`: Defaults to `http://localhost:8081`.

`SIMPLE_APM_COOKIE_TTL`: Defaults to `360` seconds (5 minutes).

Example:
```
SIMPLE_APM_PROXIED_SERVICE=http://localhost:8082  SIMPLE_APM_COOKIE_TTL=3600  SIMPLE_APM_PORT=8081 ./simple-apm-start-allow-service
```