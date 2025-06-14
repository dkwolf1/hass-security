>
See [docker/example.hubspoke.docker-compose.yml](https://github.com/hass-security/hass-security/blob/master/docker/example.hubspoke.docker-compose.yml)
for a docker-compose file.

> The following guide was contributed by @TinJoy59 in #417
> It describes how to deploy the Hass-Security in Hub/Spoke mode, where the Hub is running in Docker, and the Spokes (
> collectors) are running as binaries.
> He's using Proxmox & Synology in his guide, however this should be applicable for almost anyone

# S.M.A.R.T. Monitoring with Hass-Security across machines

![drawing-3-1671744407](https://user-images.githubusercontent.com/86809766/209230023-bf1ef9f8-65c4-454e-9e1a-be1293cd737e.png)

### 🤔 The problem:

Hass-Security offers a nice Docker package called "Omnibus" that can monitor HDDs attached to a Docker host with relative
ease. Hass-Security can also be installed in a Hub-Spoke layout where Web interface, Database and Collector come in 3
separate packages. The official documentation assumes that the spokes in the "Hub-Spokes layout" run Docker, which is
not always the case. The third approach is to install Hass-Security manually, entirely outside of Docker.

### 💡 The solution:

This tutorial provides a hybrid configuration where the Hub lives in a Docker instance while the spokes have only
Hass-Security Collector installed manually. The Collector periodically send data to the Hub. It's not mind-boggling hard to
understand but someone might struggle with the setup. This is for them.

### 🖥️ My setup:

I have a Proxmox cluster where one VM runs Docker and all monitoring services - Grafana, Prometheus, various exporters,
InfluxDB and so forth. Another VM runs the NAS - OpenMediaVault v6, where all hard drives reside. The Hass-Security Collector
is triggered every 30min to collect data on the drives. The data is sent to the Docker VM, running InfluxDB.

## Setting up the Hub

![drawing-3-1671744714](https://user-images.githubusercontent.com/86809766/209230113-c954d834-521b-4555-bcd2-eb6b80f343be.png)

The Hub consists of Hass-Security Web - a web interface for viewing the SMART data. And InfluxDB, where the smartmon data is
stored.

[🔗This is the official Hub-Spoke layout in docker-compose.](https://github.com/hass-security/hass-security/blob/master/docker/example.hubspoke.docker-compose.yml)
We are going to reuse parts of it. The ENV variables provide the necessary configuration for the initial setup, both for
InfluxDB and Hass-Security.

If you are working with and existing InfluxDB instance, you can forgo all the `INIT` variables as they already exist.

The official Hass-Security documentation has a
sample [hass-security.yaml ](https://github.com/hass-security/hass-security/blob/master/example.hass-security.yaml)file that normally
contains the connection and notification details but I always find it easier to configure as much as possible in the
docker-compose.

```yaml
version: "3.4"

networks:
  monitoring: # A common network for all monitoring services to communicate into
    external: true
  notifications: # To Gotify or another Notification service
    external: true

services:
  influxdb:
    container_name: influxdb
    image: influxdb:2.1-alpine
    ports:
      - 8086:8086
    volumes:
      - ${DIR_CONFIG}/influxdb2/db:/var/lib/influxdb2
      - ${DIR_CONFIG}/influxdb2/config:/etc/influxdb2
    environment:
      - DOCKER_INFLUXDB_INIT_MODE=setup
      - DOCKER_INFLUXDB_INIT_USERNAME=Admin
      - DOCKER_INFLUXDB_INIT_PASSWORD=${PASSWORD}
      - DOCKER_INFLUXDB_INIT_ORG=homelab
      - DOCKER_INFLUXDB_INIT_BUCKET=hass-security
      - DOCKER_INFLUXDB_INIT_ADMIN_TOKEN=your-very-secret-token
    restart: unless-stopped
    networks:
      - monitoring

  hass-security:
    container_name: hass-security
    image: ghcr.io/hass-security/hass-security:master-web
    ports:
      - 8080:8080
    volumes:
      - ${DIR_CONFIG}/hass-security/config:/opt/hass-security/config
    environment:
      - SCRUTINY_WEB_INFLUXDB_HOST=influxdb
      - SCRUTINY_WEB_INFLUXDB_PORT=8086
      - SCRUTINY_WEB_INFLUXDB_TOKEN=your-very-secret-token
      - SCRUTINY_WEB_INFLUXDB_ORG=homelab
      - SCRUTINY_WEB_INFLUXDB_BUCKET=hass-security
      # Optional but highly recommended to notify you in case of a problem
      - SCRUTINY_NOTIFY_URLS=["http://gotify:80/message?token=a-gotify-token"]
    depends_on:
      - influxdb
    restart: unless-stopped
    networks:
      - notifications
      - monitoring
```

A freshly initialized Hass-Security instance can be accessed on port 8080, eg. `192.168.0.100:8080`. The interface will be
empty because no metrics have been collected yet.

## Setting up a Spoke ***without*** Docker

![drawing-3-1671744208](https://user-images.githubusercontent.com/86809766/209230155-386a8644-b506-497f-8245-0d24e15c9063.png)

A spoke consists of the Hass-Security Collector binary that is run on a set interval via crontab and sends the data to the
Hub. The official
documentation [describes the manual setup of the Collector](https://github.com/hass-security/hass-security/blob/master/docs/INSTALL_MANUAL.md#collector)
- dependencies and step by step commands. I have a shortened version that does the same thing but in one line of code.

```bash
# Installing dependencies
apt install smartmontools -y 

# 1. Create directory for the binary
# 2. Download the binary into that directory
# 3. Make it exacutable
# 4. List the contents of the library for confirmation
mkdir -p /opt/hass-security/bin && \
curl -L https://github.com/hass-security/hass-security/releases/download/v0.8.1/hass-security-collector-metrics-linux-amd64 > /opt/hass-security/bin/hass-security-collector-metrics-linux-amd64 && \
chmod +x /opt/hass-security/bin/hass-security-collector-metrics-linux-amd64 && \
ls -lha /opt/hass-security/bin
```

<p class="callout warning">When downloading Github Release Assests, make sure that you have the correct version. The provided example is with Release v0.5.0. [The release list can be found here.](https://github.com/hass-security/hass-security/releases) </p>

Once the Collector is installed, you can run it with the following command. Make sure to add the correct address and
port of your Hub as `--api-endpoint`.

```bash
/opt/hass-security/bin/hass-security-collector-metrics-linux-amd64 run --api-endpoint "http://192.168.0.100:8080"
```

This will run the Collector once and populate the Web interface of your Hass-Security instance. In order to collect metrics
for a time series, you need to run the command repeatedly. Here is an example for crontab, running the Collector every
15min.

```bash
# open crontab
crontab -e

# add a line for Hass-Security
*/15 * * * * /opt/hass-security/bin/hass-security-collector-metrics-linux-amd64 run --api-endpoint "http://192.168.0.100:8080"
```

The Collector has its own independent config file that lives in `/opt/hass-security/config/collector.yaml` but I did not find
a need to modify
it. [A default collector.yaml can be found in the official documentation.](https://github.com/hass-security/hass-security/blob/master/example.collector.yaml)

## Setting up a Spoke ***with*** Docker

![drawing-3-1671744277](https://user-images.githubusercontent.com/86809766/209230176-87c9e55a-4e3e-4f5f-9609-335d41529f3d.png)

Setting up a remote Spoke in Docker requires you to split
the [official Hub-Spoke layout docker-compose.yml](https://github.com/hass-security/hass-security/blob/master/docker/example.hubspoke.docker-compose.yml)
. In the following docker-compose you need to provide the `${API_ENDPOINT}`, in my case `http://192.168.0.100:8080`.
Also all drives that you wish to monitor need to be presented to the container under `devices`.

The image handles the periodic scanning of the drives.

```yaml
version: "3.4"

services:

  collector:
    image: 'ghcr.io/hass-security/hass-security:master-collector'
    cap_add:
      - SYS_RAWIO
    volumes:
      - '/run/udev:/run/udev:ro'
    environment:
      COLLECTOR_API_ENDPOINT: ${API_ENDPOINT}
    devices:
      - "/dev/sda"
      - "/dev/sdb"
```
