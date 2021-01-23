# Minimalist Netflow v5 to squid-log collector written in Go

The broker listens on UDP port (default 2055), accepts Netflow traffic, and by default collects records with selected metadata formatted into squid log. Login information replaces the Mac address of the device that receives from the router mikrotik.

To build the report, it uses the [screensquid](https://sourceforge.net/projects/screen-squid/) database and its part (fetch.pl) for parsing and loading the squid log into the database

## Usage

Clone repository

`git clone https://github.com/Rid-lin/gonsquid.git`

`cd gonsquid`

Copy folder assets to /usr/share/gonsquid/

`cp /assets /usr/share/gonsquid/`

Build programm:

`make -f make_linux`

Move binary file

`mv ./bin/linux/gonsquid /usr/local/bin/`

Edit file /usr/share/gonsquid/assets/gonsquid.service

`nano /usr/share/gonsquid/assets/gonsquid.service`

E.g.

`/usr/local/bin/gonsquid -subnet=10.0.0.0/8 -subnet=192.168.0.0/16 -ignorlist=10.0.0.2 -ignorlist=:3128 -ignorlist=8.8.8.8:53 -ignorlist=ff02:: -loglevel=debug -log=/var/log/gonsquid/access.log -mtaddr=192.168.1.1:8728 -u=mikrotik_user -p=mikrotik_user_password -sqladdr=mysql_user_name:mysql_password@/screensquid`

and move to /lib/systemd/system

`mv /usr/share/gonsquid/assets/gonsquid.service /lib/systemd/system`

Make sure the log folder exists, If not then

`mkdir -p /var/log/gonsquid/`

Configuring sistemd to automatically start the program

`systemctl daemon-reload`

`systemctl start gonsquid`

`systemctl enable gonsquid`

Edit the file "fetch.pl" in accordance with the comments and recommendations.

Add a task to cron (start every 5 minutes)

`crontab -e`

`*/5 * * * * /var/www/screensquid/fetch.pl > /dev/null 2>&1`

## Supported command line parameters

```
Usage of gonsquid.exe:
  -addr string
        Address and port to listen NetFlow packets (default "0.0.0.0:2055")
  -buffer int
        Size of RxQueue, i.e. value for SO_RCVBUF in bytes (default 212992)
  -gmt string
        GMT offset time (default "+0500")
  -ignorlist value
        List of lines that will be excluded from the final log
  -interval string
        Interval to getting info from Mikrotik (default "10m")
  -log string
        The file where logs will be written in the format of squid logs
  -loglevel string
        Log level (default "info")
  -m4maddr string
        Listen address for response mac-address from mikrotik (default ":3030")
  -mtaddr string
        The address of the Mikrotik router, from which the data on the comparison of the MAC address and IP address is taken
  -p string
        The password of the user of the Mikrotik router, from which the data on the comparison of the mac-address and IP-address is taken
  -sqladdr string
        string to connect DB (e.g. username:password@protocol(address)/dbname?param=value) More details in https://github.com/go-sql-driver/mysql#dsn-data-source-name
  -subnet value
        List of subnets traffic between which will not be counted
  -tls
        Using TLS to connect to a router
  -u string
        User of the Mikrotik router, from which the data on the comparison of the MAC address and IP address is taken
```

## Credits

This project was created with help of:

* https://github.com/strzinek/gonflux
* https://sourceforge.net/projects/screen-squid/