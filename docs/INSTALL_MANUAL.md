# Manual Install

While the easiest way to get started with [Hass-Security is using Docker](https://github.com/hass-security/hass-security#docker),
it is possible to run it manually without much work. You can even mix and match, using Docker for one component and
a manual installation for the other. There's also [an installer](INSTALL_ANSIBLE.md) which automates this manual installation procedure.

Hass-Security is made up of three components: an influxdb Database, a collector and a webapp/api. Here's how each component can be deployed manually.

> Note: the `/opt/hass-security` directory is not hardcoded, you can use any directory name/path.

## InfluxDB

Please follow the official InfluxDB installation guide. Note, you'll need to install v2.2.0+. 

https://docs.influxdata.com/influxdb/v2.2/install/

## Webapp/API

### Dependencies

Since the webapp is packaged as a stand alone binary, there isn't really any software you need to install other than `glibc`
which is included by most linux OS's already.


### Directory Structure

Now let's create a directory structure to contain the Hass-Security files & binary.

```
mkdir -p /opt/hass-security/config
mkdir -p /opt/hass-security/web
mkdir -p /opt/hass-security/bin
```

### Config file

While it is possible to run the webapp/api without a config file, the defaults are designed for use in a container environment,
and so will need to be overridden. So the first thing you'll need to do is create a config file that looks like the following:

```
# stored in /opt/hass-security/config/hass-security.yaml

version: 1

web:
  database:
    # The Hass-Security webapp will create a database for you, however the parent directory must exist.
    location: /opt/hass-security/config/hass-security.db
  src:
    frontend:
      # The path to the Hass-Security frontend files (js, css, images) must be specified.
      # We'll populate it with files in the next section
      path: /opt/hass-security/web
  
  # if you're runnning influxdb on a different host (or using a cloud-provider) you'll need to update the host & port below. 
  # token, org, bucket are unnecessary for a new InfluxDB installation, as Hass-Security will automatically run the InfluxDB setup, 
  # and store the information in the config file. If you 're re-using an existing influxdb installation, you'll need to provide
  # the `token`
  influxdb:
    host: localhost
    port: 8086
#    token: 'my-token'
#    org: 'my-org'
#    bucket: 'bucket'
```

> Note: for a full list of available configuration options, please check the [example.hass-security.yaml](https://github.com/hass-security/hass-security/blob/master/example.hass-security.yaml) file.

### Download Files

Next, we'll download the Hass-Security API binary and frontend files from the [latest Github release](https://github.com/hass-security/hass-security/releases).
The files you need to download are named:

- **hass-security-web-linux-amd64** - save this file to `/opt/hass-security/bin`
- **hass-security-web-frontend.tar.gz** - save this file to `/opt/hass-security/web`

### Prepare Hass-Security

Now that we have downloaded the required files, let's prepare the filesystem.

```
# Let's make sure the Hass-Security webapp is executable.
chmod +x /opt/hass-security/bin/hass-security-web-linux-amd64

# Next, lets extract the frontend files.
# NOTE: after extraction, there **should not** be a `dist` subdirectory in `/opt/hass-security/web` directory.
cd /opt/hass-security/web
tar xvzf hass-security-web-frontend.tar.gz --strip-components 1 -C .


# Cleanup
rm -rf hass-security-web-frontend.tar.gz
```

### Start Hass-Security Webapp

Finally, we start the Hass-Security webapp:

```
/opt/hass-security/bin/hass-security-web-linux-amd64 start --config /opt/hass-security/config/hass-security.yaml
```

The webapp listens for traffic on `http://0.0.0.0:8080` by default.


## Collector

### Dependencies

Unlike the webapp, the collector does have some dependencies:

- `smartctl`, v7+
- `cron` (or an alternative process scheduler)

Unfortunately the version of `smartmontools` (which contains `smartctl`) available in some of the base OS repositories is ancient.
So you'll need to install the v7+ version using one of the following commands:

- **Ubuntu (22.04/Jammy/LTS):** `apt-get install -y smartmontools`
- **Ubuntu (18.04/Bionic):** `apt-get install -y smartmontools=7.0-0ubuntu1~ubuntu18.04.1`
- **Centos8:**
    - `dnf install https://extras.getpagespeed.com/release-el8-latest.rpm`
    - `dnf install smartmontools`
- **FreeBSD:** `pkg install smartmontools`

### Directory Structure

Now let's create a directory structure to contain the Hass-Security collector binary.

```
mkdir -p /opt/hass-security/bin
```


### Download Files

Next, we'll download the Hass-Security collector binary from the [latest Github release](https://github.com/hass-security/hass-security/releases).
The file you need to download is named:

- **hass-security-collector-metrics-linux-amd64** - save this file to `/opt/hass-security/bin`


### Prepare Hass-Security

Now that we have downloaded the required files, let's prepare the filesystem.

```
# Let's make sure the Hass-Security collector is executable.
chmod +x /opt/hass-security/bin/hass-security-collector-metrics-linux-amd64
```

### Start Hass-Security Collector, Populate Webapp

Next, we will manually trigger the collector, to populate the Hass-Security dashboard:

> NOTE: if you need to pass a config file to the hass-security collector, you can provide it using the `--config` flag.

```
/opt/hass-security/bin/hass-security-collector-metrics-linux-amd64 run --api-endpoint "http://localhost:8080"
```

### Schedule Collector with Cron

Finally you need to schedule the collector to run periodically.
This may be different depending on your OS/environment, but it may look something like this:

```
# open crontab
crontab -e

# add a line for Hass-Security
*/15 * * * * . /etc/profile; /opt/hass-security/bin/hass-security-collector-metrics-linux-amd64 run --api-endpoint "http://localhost:8080"
```
