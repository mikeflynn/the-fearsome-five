```
 _______ __   __ _______   _______ _______ _______ ______   _______ _______ __   __ _______   _______ 
|       |  | |  |       | |       |       |   _   |    _ | |       |       |  |_|  |       | |       |
|_     _|  |_|  |    ___| |    ___|    ___|  |_|  |   | || |  _____|   _   |       |    ___| |   ____|
  |   | |       |   |___  |   |___|   |___|       |   |_||_| |_____|  | |  |       |   |___  |  |____ 
  |   | |       |    ___| |    ___|    ___|       |    __  |_____  |  |_|  |       |    ___| |_____  |
  |   | |   _   |   |___  |   |   |   |___|   _   |   |  | |_____| |       | ||_|| |   |___   _____| |
  |___| |__| |__|_______| |___|   |_______|__| |__|___|  |_|_______|_______|_|   |_|_______| |_______|
```

A post exploitation command and control system.

Written in golang in three parts: server, client, and a cli admin.

# THIS IS STILL IN ACTIVE DEVELOPMENT AND NOT YET READY FOR A RELEASE. IF YOU ARE INTERESTED IN HELPING, PLEASE SUBMIT A PR OR REACH OUT TO @thatmikeflynn ON TWITTER.

## Why?

There are many C2s already out in the world written by hackers and others in the infosec field, but I wanted to give it a try to understand the concept better, and hopefully make something that is functional and looks cool and cyberpunk enough to feel at home on a terrible hacker movie.

## Features

* Separate server and admin applications to allow for better command concealment.
* Client communication is done over websockets, which are easy to connect to through NAT and can look like standard web traffic at first glance.
* Send commands to clients.
* Send and retrieve files.
* Send Powershell commands to Windows hosts.
* The ability to easily create new admin applications that interface with the server's simple REST API interface.
* Plugin architecture (future)

## Demo

-- VIDEO DEMO ONCE 1.0 IS READY --

## Usage

### Client

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

### Server

```
❯ ./server/server -h
Usage of ./server/server:
  -listen string
      API listen address. (default "0.0.0.0:8000")
  -verbose
      Display extra logging.
```

### Admin

Coming soon!

## Development

A Docker environment is included in the repo for both the  client and the server, and they can be started together via the `docker-compose.yml` file in the root.

The way to start a server with three clients is with a command like this: `docker-compose up --scale client=3`
