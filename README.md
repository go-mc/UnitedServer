# UnitedServer

Switching between server when you are playing Minecraft!

- Support graceful shutdown
- Switch server with single command

By default, this program listen at tcp:25566, then proxy all connection to localhost:25565.
After players join game, they can use `/connect <addr>` command to switch to other any server.

## Todo list
- Online-mode support
- Whitelist and Blacklist
