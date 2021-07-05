add dont-require-permissions=no name=leaseprint owner=admin policy=ftp,reboot,read,write,policy,test,password,sniff,sensitive,romon source="/ip dhcp-server lease \r\
    \n:foreach i in=[/ip dhcp-server lease find] do={\r\
    \n:local activeAddress [get \$i active-address]\r\
    \n:local addressL [get \$i address]\r\
    \n:local allowDualStackQueue [get \$i allow-dual-stack-queue]\r\
    \n:local clientId [get \$i client-id]\r\
    \n:local disabledL [get \$i disabled]\r\
    \n:local insertQueueBefore [get \$i insert-queue-before]\r\
    \n:local radiusL [get \$i radius]\r\
    \n:local statusL [get \$i status]\r\
    \n:local activeClientId [get \$i active-client-id]\r\
    \n:local addressLists [get \$i address-lists]\r\
    \n:local alwaysBroadcast [get \$i always-broadcast]\r\
    \n:local commentL [get \$i comment]\r\
    \n:local dynamicL [get \$i dynamic]\r\
    \n:local lastSeen [get \$i last-seen]\r\
    \n:local rateLimit [get \$i rate-limit]\r\
    \n:local useSrcMac [get \$i use-src-mac]\r\
    \n:local activeMacAddress [get \$i active-mac-address]\r\
    \n:local agentCircuitId [get \$i agent-circuit-id]\r\
    \n:local blockAccess [get \$i block-access]\r\
    \n:local dhcpOption [get \$i dhcp-option]\r\
    \n:local expiresAfter [get \$i expires-after]\r\
    \n:local leaseTime [get \$i lease-time]\r\
    \n:local serverL [get \$i server]\r\
    \n:local activeServer [get \$i active-server]\r\
    \n:local agentRemoteId [get \$i agent-remote-id]\r\
    \n:local blockedL  [get \$i blocked ]\r\
    \n:local dhcpOptionSet [get \$i dhcp-option-set]\r\
    \n:local hostName [get \$i host-name]\r\
    \n:local macAddress [get \$i mac-address]\r\
    \n:local srcMacAddress [get \$i src-mac-address]\r\
    \n:put \"activeAddress='\$activeAddress';addressL='\$addressL';allowDualStackQueue='\$allowDualStackQueue';clientId='\$clientId';disabledL='\$disabledL';insertQueueBefore='\
    \$insertQueueBefore';radiusL='\$radiusL';statusL='\$statusL';activeClientId='\$activeClientId';addressLists='\$addressLists';alwaysBroadcast='\$alwaysBroadcast';commentL='\$\
    commentL';dynamicL='\$dynamicL';lastSeen='\$lastSeen';rateLimit='\$rateLimit';useSrcMac='\$useSrcMac';activeMacAddress='\$activeMacAddress';agentCircuitId='\$agentCircuitId'\
    ;blockAccess='\$blockAccess';dhcpOption='\$dhcpOption';expiresAfter='\$expiresAfter';leaseTime='\$leaseTime';serverL='\$serverL';activeServer='\$activeServer';agentRemoteId=\
    '\$agentRemoteId';blockedL='\$blockedL';dhcpOptionSet='\$dhcpOptionSet';hostName='\$hostName';macAddress='\$macAddress';srcMacAddress='\$srcMacAddress'\"\r\
    \n}"
