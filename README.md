# moody-core

This repository contains the main engine of the moody architecture. 

You can use it to interface with sensors communicating via MQTT or 
running firmware based on Moody-frameworks, like [MoodyNodeEsp](https://github.com/antima/MoodyNodeEsp).

- [moody-core](#moody-core)
- [Build from source](#build-from-source)

# Build from source

The following command will build the application into the root directory of the repo.

```bash
mage build
```

To install moody globally and run it as a systemd service, after buiding, run:

```bash
sudo mage install
```

