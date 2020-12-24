# Minimalist Netflow v5 to squid-log collector written in Go 

The broker listens on UDP port (default 2055), accepts Netflow traffic, and by default collects records with selected metadata formatted in line protocol into squid log. Login information replaces the Mac address of the device that receives from the router mikrotik.

## Usage

```
goflow -subnet=10.0.0.0/8 -subnet=192.168.0.0/16 -ignorlist=10.0.0.2 -ignorlist=:3128 -ignorlist=8.8.8.8:53 -ignorlist=ff02:: -loglevel=debug -log=/var/log/flow/access.log -mtaddr=192.168.1.1:8728 -u=user -p=password
  
```

Supported command line parameters:
```
  -addr string
        Address and port to listen NetFlow packets (default "0.0.0.0:2055")
  -buffer int
        Size of RxQueue, i.e. value for SO_RCVBUF in bytes (default 212992)
  -gmt string
        GMT offset time (default "+0500")
  -ignorlist value
        List of lines that will be excluded from the final log
  -interval string
        Interval to getting info from Mikrotik in minute (default "10")
  -log string
        The file where logs will be written in the format of squid logs
  -loglevel string
        Log level (default "info")
  -m4maddr string
        Listen address for  (default "localhost:3030")
  -mtaddr string
        The address of the Mikrotik router, from which the data on the comparison of the MAC address and IP address is taken
  -p string
        The password of the user of the Mikrotik router, from which the data on the comparison of the mac-address and IP-address is taken
  -port string
        Address for service mac-address determining (default "localhost:3030")
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
