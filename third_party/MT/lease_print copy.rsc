/ip dhcp-server lease
:foreach i in=[/ip dhcp-server lease find] do={[:local ID [get $i .id];:put ".id=$ID"];}
