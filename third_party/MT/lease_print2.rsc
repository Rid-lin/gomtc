/ip dhcp-server lease
:foreach i in=[/ip dhcp-server lease find] do={
    [:local ID [get $i id];
    [:local activeAddress [get $i active-address]];
    [:local addressL [get $i address]];
    [:local allowDualStackQueue [get $i allow-dual-stack-queue]];
    [:local clientId [get $i client-id]];
    [:local disabledL [get $i disabled]];
    [:local insertQueueBefore [get $i insert-queue-before]];
    [:local radiusL [get $i radius]];
    [:local statusL [get $i status]];
    [:local activeClientId [get $i active-client-id]];
    [:local addressLists [get $i address-lists]];
    [:local alwaysBroadcast [get $i always-broadcast]];
    [:local commentL [get $i comment]];
    [:local dynamicL [get $i dynamic]];
    [:local lastSeen [get $i last-seen]];
    [:local rateLimit [get $i rate-limit]];
    [:local useSrcMac [get $i use-src-mac]];
    [:local activeMacAddress [get $i active-mac-address]];
    [:local agentCircuitId [get $i agent-circuit-id]];
    [:local blockAccess [get $i block-access]];
    [:local dhcpOption [get $i dhcp-option]];
    [:local expiresAfter [get $i expires-after]];
    [:local leaseTime [get $i lease-time]];
    [:local serverL [get $i server]];
    [:local activeServer [get $i active-server]];
    [:local agentRemoteId [get $i agent-remote-id]];
    [:local blockedL  [get $i blocked ]];
    [:local dhcpOptionSet [get $i dhcp-option-set]];
    [:local hostName [get $i host-name]];
    [:local macAddress [get $i mac-address]];
    [:local srcMacAddress [get $i src-mac-address]];
    :put ".id=$ID;active-address=$activeAddress;address=$addressL;allow-dual-stack-queue=$allowDualStackQueue;client-id=$clientId;disabled=$disabledL;insert-queue-before=$insertQueueBefore;radius=$radiusL;status=$statusL;active-client-id=$activeClientId;address-lists=$addressLists;always-broadcast=$alwaysBroadcast;comment=$commentL;dynamic=$dynamicL;last-seen=$lastSeen;rate-limit=$rateLimit;use-src-mac=$useSrcMac;active-mac-address=$activeMacAddress;agent-circuit-id=$agentCircuitId;block-access=$blockAccess;dhcp-option=$dhcpOption;expires-after=$expiresAfter;lease-time=$leaseTime;server=$serverL;active-server=$activeServer;agent-remote-id=$agentRemoteId;blocked=$blockedL;dhcp-option-set=$dhcpOptionSet;host-name=$hostName;mac-address=$macAddress;src-mac-address=$srcMacAddress"];}
