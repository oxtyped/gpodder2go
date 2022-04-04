# gpodder2go

gpodder2go is a simple self-hosted, golang, drop-in replacement for gpodder/mygpo server to handle podcast subscriptions management for gpodder clients.

## Goal

- To build an easily deployable and private self-hosted drop-in replacement for gpodder.net to facilitate private and small group sychronization of podcast subscriptions with fediverse support

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

- To join gpodder2go with the fediverse to allow for independent gpodder2go servers to communicate with one another to discover and share like-minded podcasts that the communities are listening to

### Non-goals

gpodder2go will not come with it a web frontend and will solely be an API server. While this is not totally fixed and may change in the future, the current plan is to not handle anything frontend.
