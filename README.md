[![Go Report Card](https://goreportcard.com/report/git.vegner.org/vsvegner/gomtc)](https://goreportcard.com/report/git.vegner.org/vsvegner/gomtc)

# gomtc

[gomtc](https://git.vegner.org/vsvegner/gomtc) is a Squid log analyzer with a web interface and the ability to control the device via ssh Mikrotik is written in Go
Reads logs from the access.log file and saves them as a sqlite database.

## Install

Clone repository

`git clone https://git.vegner.org/vsvegner/gomtc.git`

`cd gomtc`

Copy folder assets to /usr/share/gomtc/

`cp /assets /usr/share/gomtc/`

Copy the config folder to /etc/gomtc

`cp /config /etc/gomtc/`

Build programm:

`make build`

Move binary file

`mv ./bin/gomtc_linux_amd64 /usr/local/bin/`

Edit file /usr/share/gomtc/assets/gomtc.service

`nano /usr/share/gomtc/assets/gomtc.service`

E.g.

`/usr/local/bin/gomtc -subnet=10.0.0.0/8 -subnet=192.168.0.0/16 -ignorlist=10.0.0.2 -ignorlist=:3128 -ignorlist=8.8.8.8:53 -ignorlist=ff02:: -loglevel=debug -log=/var/log/gomtc/access.log -mtaddr=192.168.1.1:8728 -u=mikrotik_user -p=mikrotik_user_password -sqladdr=mysql_user_name:mysql_password@/screensquid`

and move to /lib/systemd/system

`mv /usr/share/gomtc/assets/gomtc.service /lib/systemd/system`

Make sure the log folder exists, If not then

`mkdir -p /var/log/gomtc/`

Configuring sistemd to automatically start the program

`systemctl daemon-reload`

`systemctl start gomtc`

`systemctl enable gomtc`

## Usage

```usage
Usage:
  -assets_path string
        The path to the assets folder where the template files are located (default "/etc/gomtc/assets")
  -block_group string
        The name of the address list in MicrotiK with which access is blocked to users who have exceeded the quota. (default "Block")
  -config_path string
        folder path to all config files (default "/etc/gomtc")
  -csv string
        Output to csv (default "false")
  -default_quota_daily string
        Default daily traffic consumption quota (default "0")
  -default_quota_hourly string
        Default hourly traffic consumption quota (default "0")
  -default_quota_monthly string
        Default monthly traffic consumption quota (default "0")
  -devices_retry_delay string
        Interval to getting info from Mikrotik (default "1m")
  -flow_addr string
        Address and port to listen NetFlow packets (default "0.0.0.0:2055")
  -friends string
        List of aliases, IP addresses, friends' logins
  -gonsquid_addr string
        Listen address for HTTP-server (default ":3030")
  -ignor_list string
        List of lines that will be excluded from the final log
  -listen_addr string
        Listen address for HTTP-server (default ":3031")
  -loc string
        Location for time (default "Asia/Yekaterinburg")
  -log_level string
        Log level: panic, fatal, error, warn, info, debug, trace (default "info")
  -log_path string
        folder path to logs-file (default "/var/log/gomtc")
  -max_ssh_retries string
        The number of attempts to connect to the microtik router (default "-1")
  -mt_addr string
        The address of the Mikrotik router, from which the data on the comparison of the MAC address and IP address is taken
  -mt_pass string
        The password of the user of the Mikrotik router, from which the data on the comparison of the mac-address and IP-address is taken
  -mt_user string
        User of the Mikrotik router, from which the data on the comparison of the MAC address and IP address is taken
  -name_file_to_log string
        The file where logs will be written in the format of squid logs
  -no_flow string
        When this parameter is specified, the netflow packet listener is not launched, therefore, log files are not created, but only parsed. (default "true")
  -parse_delay string
        Delay parsing logs (default "10m")
  -receive_buffer_size_bytes string
        Size of RxQueue, i.e. value for SO_RCVBUF in bytes
  -size_one_kilobyte string
        The number of bytes in one megabyte (default "1024")
  -ssh_port string
        The port of the Mikrotik router for SSH connection (default "22")
  -ssh_retry_delay string
        Interval of attempts to connect to MT (default "0")
  -sub_nets string
        List of subnets traffic between which will not be counted
  -use_tls string
        Using TLS to connect to a router (default "false")
```

## Credits

This project was created with help of:

* <https://github.com/strzinek/gonflux>
* <https://sourceforge.net/projects/screen-squid/>
