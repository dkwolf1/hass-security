# pfsense Install

This bascially follows the [Manual collector instructions](https://github.com/hass-security/hass-security/blob/master/docs/INSTALL_MANUAL.md#collector) and assumes you are running a hub and spoke deployment and already have the web app setup.


### Dependencies

SSH into pfsense, hit `8` for the shell and install the required dependencies.

```
pkg install smartmontools
```

Ensure smartmontools is v7+. This won't be a problem in pfsense 2.6.0+


### Directory Structure

Now let's create a directory structure to contain the Hass-Security collector binary.

```
mkdir -p /opt/hass-security/bin
```


### Download Files

Next, we'll download the Hass-Security collector binary from the [latest Github release](https://github.com/hass-security/hass-security/releases).

> NOTE: Ensure you have the latest version in the below command

```
fetch -o /opt/hass-security/bin https://github.com/hass-security/hass-security/releases/download/vX.X.X/hass-security-collector-metrics-freebsd-amd64
```


### Prepare Hass-Security

Now that we have downloaded the required files, let's prepare the filesystem.

```
chmod +x /opt/hass-security/bin/hass-security-collector-metrics-freebsd-amd64
```


### Start Hass-Security Collector, Populate Webapp

Next, we will manually trigger the collector, to populate the Hass-Security dashboard:

> NOTE: if you need to pass a config file to the hass-security collector, you can provide it using the `--config` flag.

```
/opt/hass-security/bin/hass-security-collector-metrics-freebsd-amd64 run --api-endpoint "http://localhost:8080"
```
> NOTE: change the IP address to that of your web app

### Schedule Collector with Cron

Finally you need to schedule the collector to run periodically.

Login to the pfsense webGUI and head to `Services/Cron` add an entry with the following details:

```
Minute: */15
Hour: *
Day of the Month: *
Month of the Year: *
Day of the Week: *
User: root
Command: /opt/hass-security/bin/hass-security-collector-metrics-freebsd-amd64 run --api-endpoint "http://localhost:8080" >/dev/null 2>&1
```
> NOTE: `>/dev/null 2>&1` is used to stop cron confirmation emails being sent.
