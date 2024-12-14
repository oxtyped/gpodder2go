# gpodder2go

gpodder2go is a simple self-hosted, golang, drop-in replacement for gpodder/mygpo server to handle podcast subscriptions management for gpodder clients.

## Goal

To build an easily deployable and private self-hosted drop-in replacement for gpodder.net to facilitate private and small group sychronization of podcast subscriptions with fediverse support

### Current Goal

- To support the authentication and storing/syncing of subscriptions and episode actions on multi-devices accounts
  - Target to fully support the following [gpodder APIs](https://gpoddernet.readthedocs.io/en/latest/api/index.html)
    - Authentication API
    - Subscriptions API
    - Episode Actions API
    - Device API
    - Device Synchronization API
- To provide a pluggable interface to allow developers to pick and choose the data stores that they would like to use (file/in-memory/rdbms)

### Stretch Goal

To join gpodder2go with the fediverse to allow for independent gpodder2go servers to communicate with one another to discover and share like-minded podcasts that the communities are listening to

### Non-goals

gpodder2go will not come with it a web frontend and will solely be an API server. While this is not totally fixed and may change in the future, the current plan is to not handle anything frontend.

### Database Requirement

gpodder2go requires a database to manage the subscription and user states. Currently the project only supports SQLite with plans to support other databases. The current database mechanism is managed by a [DataInterface](https://github.com/oxtyped/gpodder2go/blob/main/pkg/data/types.go#L8-L21) which allows for quick easy support of new database stores when needed.

### Quickstart

1. Download the [respective binary](https://github.com/oxtyped/gpodder2go/releases)
2. Initialize the necessary database and configurations

```
$ ./gpodder2go init
```

4. Start the gpodder server
```
$ VERIFIER_SECRET_KEY="" ./gpodder2go serve
```

**Note**: `VERIFIER_SECRET_KEY` is a required env var. This value will be used to sign and verify the sessionid which will be used to authenticate users.

5. Create a new user
```
$ gpodder2go accounts create <username> --email="<email>" --name="<display_name>" --password="<password>"
```
**Note**: Each of the commands have a bunch of flags that you can use, to view the full list of available flags, use `--help` or `-h` after the commands.

### Limitations

Right now it appears that the gpodder client doesn't fully support auth (see: https://github.com/gpodder/gpodder/issues/617 and https://github.com/gpodder/gpodder/issues/1358) even though the specification (https://gpoddernet.readthedocs.io/en/latest/api/reference/auth.html) explicitly defines it.

In order to allow gpodder client access to the gpodder server, please run `gpodder2go` in non-auth mode.

```
$ gpodder2go serve --no-auth
```

**Note**: This will allow anyone with access to retrieve your susbcriptions data and list. Please take the necessary steps to secure your instance and data.

Alternatively, you can switch to use [Antennapod](https://antennapod.org/) which has implemented the login spec which gpodder2go currently supports.

### Supported Clients

#### [Antennapod](https://antennapod.org/)

These features are all working with Antennapod:
    - Authentication API
    - Subscriptions API
    - Episode Actions API
    - Device API
    - Device Synchronization API

To start using with two devices, especially if you want to transfer state
from old_phone to new_phone:

#. Log in with old_phone and force a full sync.
#. Log in with new_phone, but select old_phone as the device. Subscriptions will sync.
#. Log out with new_phone.
#. Log in with new_phone and create a new device ID for it.
#. Use the API (such as with curl) to [create a sync group](https://gpoddernet.readthedocs.io/en/latest/api/reference/sync.html#device-synchronization-api) with both devices.

After that, episode state will sync between them, and a new subscription on
either one will propagate to the other.

### Development

```
$ go run main.go
```

### Distribution Packages

#### Gentoo
Available with a custom overlay at:
https://github.com/seigakaku/gentoo_ebuilds/tree/master/media-sound/gpodder2go

Add with:
```
# eselect repository add seiga git https://github.com/seigakaku/gentoo_ebuilds
```

### Docker

```sh
$ docker run -d \
--name gpodder2go \
-p 3005:3005 \
-e NO_AUTH=<true or false> \
-v <data_directory>:/data \
ghcr.io/oxtyped/gpodder2go:main
```

With docker compose:

```yaml
version: '3'
services:
  gpodder2go:
    image: ghcr.io/oxtyped/gpodder2go:main
    ports:
      - 3005:3005
    environment:
      - NO_AUTH=<true or false>
    volumes:
      - ./gpodder2go:/data
    restart: unless-stopped
```

To configure the server run

```sh
$ docker exec --it gpodder2go /gpodder2go ...
```

#### Build docker image from source

Build with:

```
$ git clone https://github.com/oxtyped/gpodder2go
$ cd gpodder2go
$ docker build -t oxtyped/gpodder2go .
```

Run with:

```
$ docker run --rm -it -p 3005:3005 oxtyped/gpodder2go
```

For persistent data, you can map `/data` as a volume:

```
$ docker run --rm -it -v /gpodder2go_data:/data -p 3005:3005 oxtyped/gpodder2go
```

To add a user:

```
$ docker run --rm -it -v /gpodder2go_data:/data oxtyped/gpodder2go /gpodder2go accounts create <username> --email="<email>" --name="<display_name>" --password="<password>"
```
