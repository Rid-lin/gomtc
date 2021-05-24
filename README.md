https://goreportcard.com/report/git.vegner.org/vsvegner/gomtcя[![Go Report Card](https://goreportcard.com/report/git.vegner.org/vsvegner/gomtc)](https://goreportcard.com/report/git.vegner.org/vsvegner/gomtc)

# Minimalist Netflow v5 to squid-log collector written in Go

The broker listens on UDP port (default 2055), accepts Netflow traffic, and by default collects records with selected metadata formatted into squid log. Login information replaces the Mac address of the device that receives from the router mikrotik.

To build the report, it uses the [screensquid](https://sourceforge.net/projects/screen-squid/) database and its part (fetch.pl) for parsing and loading the squid log into the database

## Usage

Clone repository

`git clone https://github.com/Rid-lin/gomtc.git`

`cd gomtc`

Copy folder assets to /usr/share/gomtc/

`cp /assets /usr/share/gomtc/`

Build programm:

`make -f make_linux`

Move binary file

`mv ./bin/linux/gomtc /usr/local/bin/`

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


## Supported command line parameters

```
```

## Credits

This project was created with help of:

* https://github.com/strzinek/gonflux
* https://sourceforge.net/projects/screen-squid/