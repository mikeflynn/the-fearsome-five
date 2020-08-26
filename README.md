# Fearsome Five
A Command and Control System

Written in golang in three parts: server, client, and a cli admin.

## Client

```
❯ ./client/client -h
Usage of ./client/client:
  -delay int
      Delay, in seconds, before reconnection attempts. (default 300)
  -reset
      If true, it will reset use the flags and reset the config file.
  -server string
      Server hostname. (default "localhost:8000")
  -unsafe
      Turn off all discovery safe guards.
  -verbose
      Additional debugging logs.
  -workdir string
      Set the working directory (default "./")
```

## Server

```
❯ ./server/server -h
Usage of ./server/server:
  -listen string
      API listen address. (default "0.0.0.0:8000")
  -verbose
      Display extra logging.
```