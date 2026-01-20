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

### Self-hosting with an existing nginx server

Assuming you have a public facing server with TLS via letsencrypt on nginx, you can configure like so:

0. Assumptions:
   - You have built gpodder2go at `~/gpodder2go/gpodder2go`
   - You have up to date certbot TLS certificate
   - Your website is enabled in nginx

1. Reverse proxy behind nginx:
# /etc/nginx/available-sites/yoursite.com
```
server {                                                                                                                                                                                                                           
    listen 3005 ssl;                                                                                                                                                                                                               
                                                                                                                                                                                                                                   
    server_name yoursite.com;                                                                                                                                                                                                     
                                                                                                                                                                                                                                   
    ssl_certificate /etc/letsencrypt/live/yoursite.com/fullchain.pem;                                                                                                                                                             
    ssl_certificate_key /etc/letsencrypt/live/yoursite.com/privkey.pem;                                                                                                                                                           
    include /etc/letsencrypt/options-ssl-nginx.conf;                                                                                                                                                                               
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;                                                                                                                                                                                 
                                                                                                                                                                                                                                   
    location / {                                                                                                                                                                                                                   
            proxy_pass         http://127.0.0.1:3050/;                                                                                                                                                                             
            proxy_redirect     off;                                                                                                                                                                                                
                                                                                                                                                                                                                                   
            proxy_set_header   Host             $host;                                                                                                                                                                             
            proxy_set_header   X-Real-IP        $remote_addr;                                                                                                                                                                      
            proxy_set_header   X-Forwarded-For  $proxy_add_x_forwarded_for;                                                                                                                                                        
                                                                                                                                                                                                                                   
            client_max_body_size       10m;                                                                                                                                                                                        
            client_body_buffer_size    128k;                                                                                                                                                                                       
                                                                                                                                                                                                                                   
            proxy_connect_timeout      90;                                                                                                                                                                                         
            proxy_send_timeout         90;                                                                                                                                                                                         
            proxy_read_timeout         90;                                                                                                                                                                                         
                                                                                                                                                                                                                                   
            proxy_buffer_size          4k;                                                                                                                                                                                         
            proxy_buffers              4 32k;                                                                                                                                                                                      
            proxy_busy_buffers_size    64k;                                                                                                                                                                                        
            proxy_temp_file_write_size 64k;                                                                                                                                                                                        
    }                                                                                                                                                                                                                              
}                                                                                                                                                                                                                                  
```

2. A systemd service (customized to my deployment, you'll have to change some user values):
# /etc/systemd/system/gpodder2go.service
```
[Unit]                                                                                                                                                                                                                             
Description=gpodder2go Service                                                                                                                                                                                                     
After=network-online.target                                                                                                                                                                                                        
Wants=network-online.target                                                                                                                                                                                                        
                                                                                                                                                                                                                                   
[Service]                                                                                                                                                                                                                          
Type=simple                                                                                                                                                                                                                        
ExecStart=/home/youruser/gpodder2go/gpodder2go serve --addr 127.0.0.1:3050                                                                                                                                                              
WorkingDirectory=/home/youruser/gpodder2go                                                                                                                                                                                              
# comment this if using a user service                                                                                                                                                                                             
User=youruser                                                                                                                                                                                                                           
                                                                                                                                                                                                                                   
# Environment variable (intentionally left blank)                                                                                                                                                                                  
Environment=VERIFIER_SECRET_KEY=foo54bar54baz62environmental                                                                                                                                                                       
                                                                                                                                                                                                                                   
# ---- Security / sandboxing ----                                                                                                                                                                                                  
                                                                                                                                                                                                                                   
# Do not grant any extra privileges                                                                                                                                                                                                
#NoNewPrivileges=yes                                                                                                                                                                                                               
                                                                                                                                                                                                                                   
# Isolate /tmp                                                                                                                                                                                                                     
PrivateTmp=yes     
[Install]
WantedBy=default.target
```

4. During setup, you'll pick "gpodder.net" and then tick "use custom server" and login with your username and password (not name, not email; see column 2 in g2g.db for this value, or use the username you chose in step 5 above). 
 You will use the URL "https://yoursite.com:3005". The port here is from the "listen" directive of the nginx config


### Supports

- [Antennapod](https://antennapod.org/)
- [Kasts](https://apps.kde.org/kasts/)

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
