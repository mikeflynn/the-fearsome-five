# Fearsome Five
A Command and Control System

Written in golang.

## Payload

* Installs and immediately makes existence known.
* Allows connections from C&C repl with SSH key
* Allows for remote updates
* Allows for drop in to shell access.
* Starts as IPFS node to allow for distributed updates.
* Can be compiled for Windows, Linux, or macOS.

## REPL

* Command and control console interface.
* Keeps track of the announced active payloads.
* Lists active payloads and can filter the list by OS, current activity level, or bandwidth.
* Can connect to a single host for direct shell/powershell access.
* Can be compiled for Linux or macOS.