# Websocket broadcaster ![Build and Test](https://github.com/neonxp/wsbroadcast/workflows/Build%20and%20Test/badge.svg)

Simple websocket chat server

## Create channel

```http request
POST /channel
Content-Type: application/json

{
  "payload": {
    "title": "New channel"
  }
}
```

Result:

```json
{
  "id": "5e7d19b98803c90bcff53f84",
  "payload": {
    "title": "New channel"
  }
}
```

## Update channel

```http request
POST /channel/5e7d19b98803c90bcff53f84
Content-Type: application/json

{
  "payload": {
    "title": "Old channel"
  }
}
```
Result:

```json
{
  "id": "5e7d19b98803c90bcff53f84",
  "payload": {
    "title": "Old channel"
  }
}
```

## Get channel

```http request
GET /channel/5e7d19b98803c90bcff53f84
```

Result:

```json
{
  "id": "5e7d19b98803c90bcff53f84",
  "payload": {
    "title": "Old channel"
  },
  "members": [
    {"id": 1, "state": ""},
    {"id": 2, "state": ""}
  ]
}
```

## Websocket

Connect to: `/channel/5e7d19b98803c90bcff53f84/ws`
