# Reverse Proxy Support

Hass-Security is designed so that it can be used with a reverse proxy, leveraging `domain`, `port` or `path` based matching to correctly route to the Hass-Security service.

For simple `domain` and/or `port` based routing, this is easy.

If your domain:port pair is similar to `http://hass-security.example.com` or `http://localhost:54321`, just update your reverse proxy configuration
to route traffic to the Hass-Security backend, which is listening on `0.0.0.0:8080` by default.

```yaml
# default config
web:
  listen:
    port: 8080
    host: 0.0.0.0
```

However if you're using `path` based routing to differentiate your reverse proxy protected services, things become more complicated.

If you'd like to access Hass-Security using a path like: `http://example.com/hass-security/`, then we need a way to configure Hass-Security so that it
understands `http://example.com/hass-security/api/health` actually means `http://localhost:8080/api/health`.

Thankfully this can be done by changing **two** settings (both are required).

1. The webserver has a `web.listen.basepath` key
2. The collectors have a `api.endpoint` key.

## Webserver Configuration

When setting the `web.listen.basepath` key in the web config file, make sure the `basepath` key is prefixed with `/`.

```yaml
# customized webserver config
web:
  listen:
    port: 8080
    host: 0.0.0.0
    # if you're using a reverse proxy like apache/nginx, you can override this value to serve hass-security on a subpath.
    # eg. http://example.com/custombasepath/* vs http://example.com:8080
    basepath: '/custombasepath'
```

## Collector Configuration 

Here's how you can update the collector `api.endpoint` key:

```yaml
# customized collector config
api:
  endpoint: 'http://localhost:8080/custombasepath'
```

# Environmental Variables.

You may also configure these values using the following environmental variables (both are required).

- `COLLECTOR_API_ENDPOINT=http://localhost:8080/custombasepath`
- `SCRUTINY_WEB_LISTEN_BASEPATH=/custombasepath`

# Real Examples

## Caddy

1. Create a Caddyfile
    ```yaml
    # Caddyfile
    :9090
    
    # The `hass-security` text in this file must match the service name in the docker-compose file below. 
    # The `/custom/` text is the custom base path hass-security will be availble on. 
    reverse_proxy /custom/* hass-security:8080

    ```
2. Create a `docker-compose.yml` file

    ```yaml
    # docker-compose.yml
    version: '3.5'
    
    services:
      hass-security:
        container_name: hass-security
        image: ghcr.io/analogj/hass-security:master-omnibus
        cap_add:
          - SYS_RAWIO
        ports:
          - "8086:8086" # influxDB admin
        volumes:
          - /run/udev:/run/udev:ro
          - ./config:/opt/hass-security/config
          - ./influxdb:/opt/hass-security/influxdb
        devices:
          - "/dev/sda"
          - "/dev/sdb"
        environment:
          - SCRUTINY_WEB_LISTEN_BASEPATH=/custom
          - COLLECTOR_API_ENDPOINT=http://localhost:8080/custom
      caddy:
        image: caddy
        volumes:
          - ./Caddyfile:/etc/caddy/Caddyfile
        ports:
          - "9090:9090"
    ```
3. run `docker-compose up`
4. visit [http://localhost:9090/custom/web](http://localhost:9090/custom/web) - access the hass-security container via caddy reverse proxy
 
## Traefik

Assuming, that you have Traefik up and running with [AutoDiscovery Using Traefik For Docker ](https://doc.traefik.io/traefik/providers/docker/),
here is an example of a `docker-compose.yml` file, with labels to enable Traefik reverse proxy and basic auth
```yaml
version: '3.5'
services:
  hass-security:
    container_name: hass-security
    image: ghcr.io/analogj/hass-security:master-omnibus
    cap_add:
      - SYS_RAWIO
      - SYS_ADMIN
    volumes:
      - /run/udev:/run/udev:ro
      - ./config:/opt/hass-security/config
      - ./influxdb:/opt/hass-security/influxdb
    labels:
      - traefik.enable=true
      - traefik.http.routers.hass-security.rule=Host(`example.com`)
      - traefik.http.services.hass-security.loadbalancer.server.port=8080
        # 2 labels below are optional, in case you want basic auth in Traefik:
      - traefik.http.routers.hass-security.middlewares=auth
      - "traefik.http.middlewares.auth.basicauth.users=user:$$2y$$05$$G11Wm/dlWpXHENK..m8se.zxvaE8USJBp1Ws56sSCrOcwWDjsYHni"
        # Note: when used in docker-compose.yml all dollar signs in the hash need to be doubled for escaping.
        # To create user:password pair, it's possible to use this command:
        # echo $(htpasswd -nB user) | sed -e s/\\$/\\$\\$/g
    devices:
      - "/dev/sda"
      - "/dev/sdb"
      - "/dev/nvme0"
```
