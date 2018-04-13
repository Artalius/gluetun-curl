# Private Internet Access Docker (OpenVPN, Alpine)

Docker VPN client to private internet access servers based on [Alpine Linux](https://alpinelinux.org/) using [OpenVPN](https://openvpn.net/) and Unbound to connect to [Cloudflare DNS 1.1.1.1 over TLS](https://developers.cloudflare.com/1.1.1.1/dns-over-tls)

[![PIA Docker OpenVPN](https://github.com/qdm12/private-internet-access-docker/raw/master/readme/title.png)](https://hub.docker.com/r/qmcgaw/private-internet-access/)

[![Build Status](https://travis-ci.org/qdm12/private-internet-access-docker.svg?branch=master)](https://travis-ci.org/qdm12/private-internet-access-docker)

[![](https://images.microbadger.com/badges/image/qmcgaw/private-internet-access.svg)](https://microbadger.com/images/qmcgaw/private-internet-access)
[![](https://images.microbadger.com/badges/version/qmcgaw/private-internet-access.svg)](https://microbadger.com/images/qmcgaw/private-internet-access)

| Download size | Image size | RAM usage | CPU usage |
| --- | --- | --- | --- |
| ?MB | 12.9MB | ?MB | Very low |

It requires:
- A Private Internet Access **username** and **password** - [Sign up](https://www.privateinternetaccess.com/pages/buy-vpn/)
- [Docker](https://docs.docker.com/install/) installed on the host

The PIA *.ovpn* configuration files are downloaded from 
[the PIA website](https://www.privateinternetaccess.com/openvpn/openvpn.zip) 
when the Docker image is built.

Cloudflare **DNS 1.1.1.1 over TLS** is used to connect to any PIA server for multiple reasons:
- Man-in-the-middle (ISP, hacker, government) can't block you from resolving the PIA server domain name. 
    *For example, `austria.privateinternetaccess.com` maps to `185.216.34.229`*
- Man-in-the-middle (ISP, hacker, government) can't see to which server you connect nor when.
    *As the domain name are sent to 1.1.1.1 over TLS, there is no way to examine what domains you are asking to be resolved*
- Lower latency than other DNS such as Google DNS, Open DNS or your ISP DNS.

## Installation & Testing

1. Run the [**tun.sh**](https://raw.githubusercontent.com/qdm12/private-internet-access-docker/master/tun.sh) script on your host machine to ensure you have the **tun** device setup

    ```bash
    wget https://raw.githubusercontent.com/qdm12/private-internet-access-docker/master/tun.sh
    sudo chmod +x tun.sh
    ./tun.sh
    ```
    
1. Create a network to be used by this container and other containers connecting to it with:

    ```bash
    docker network create pianet
    ```

1. Create a file *auth.conf* in `/yourhostpath` (for example), with:
    - On the first line: your PIA username (i.e. `js89ds7`)
    - On the second line: your PIA password (i.e. `8fd9s239G`)
    
### Using Docker only

Run the container with (change `/yourhostpath` to your actual path, and optionally `Germany`):

```bash
docker run -d --restart=always --name=pia --cap-add=NET_ADMIN \
--device=/dev/net/tun --network=pianet \
-e REGION=Germany -v /yourhostpath/auth.conf:/auth.conf:ro \
qmcgaw/private-internet-access
```

Wait about 5 seconds for it to connect to the PIA server.
You can check with:

```bash
docker logs pia
```

You should now check it works following the [Testing section](#testing)

### Using Docker Compose

1. Download [**docker-compose.yml**](https://github.com/qdm12/private-internet-access-docker/blob/master/docker-compose.yml)
1. Edit it and change `yourpath`
1. Run the container as a daemon in the background with:

   ```bash
    docker-compose up -d
    ```

    Wait about 5 seconds for it to connect to the PIA server.
    You can check with:

    ```bash
    docker logs pia
    ```
    
1. You should now check it works following the [Testing section](#testing)

## Testing

1. Check your host IP address with:

    ```bash
    curl -s ifconfig.co
    ```

1. Run the **curl** Docker container using your *pia* container with:

    ```bash
    docker run --rm --network=container:pia tutum/curl curl -s ifconfig.co
    ```

    If the displayed IP address appears and is different that your host IP address, 
    the PIA client should fully work !

## Container launch parameters

- You can change the `REGION` environment variable to one of the [regions supported by private internet access](https://www.privateinternetaccess.com/pages/network/)
- If you know what you're doing, you can change the container name (`pia`), 
  the hostname (`piaclient`) and the network name (`pianet`)

## Connect other containers to it

Connect other Docker containers to the PIA VPN connection by adding 
`--network=container:pia` when launching them.
  
## Access ports of containers connected to the VPN container

You have to use another container acting as a Reverse Proxy such as Nginx. 

**Example**:
- *Deluge* container with name **deluge** connected to the `pia` container with `--network=container:pia`
- Deluge's WebUI runs on port TCP 8112

1. Create the Nginx configuration file *nginx.conf*:

    ```
    user  nginx;
    worker_processes  1;
    error_log  /var/log/nginx/error.log warn;
    pid        /var/run/nginx.pid;
    events {
        worker_connections  1024;
    }
    http {
        include       /etc/nginx/mime.types;
        default_type  application/octet-stream;
        log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
                          '$status $body_bytes_sent "$http_referer" '
                          '"$http_user_agent" "$http_x_forwarded_for"';
        access_log  /var/log/nginx/access.log  main;
        sendfile        on;
        keepalive_timeout  65;
        server {
            listen 80;
            location / {
                proxy_pass http://deluge:8112/;
                proxy_set_header X-Deluge-Base "/";
            }
        }
        include /etc/nginx/conf.d/*.conf;
    }
    ```

1. Run the Alpine [Nginx container](https://hub.docker.com/_/nginx) with:

    ```bash
    docker -d --restart=always --name=proxypia -p 8000:80 \
    --network=pianet --link pia:deluge \
    -v /mypathto/nginx.conf:/etc/nginx/nginx.conf:ro nginx:alpine
    ```
    
1. Access the WebUI of Deluge at [localhost:8000](http://localhost:8000)

For more containers, add more `--link pia:xxx` and modify *nginx.conf* accordingly
