BEGIN{}
{
  if ($0 ~/ipv4_address/) {
    print $0
    print "            public_net_access:"
    mLine = $0
    gsub(70,72,mLine)
    print mLine
  } else if ($0 ~/name: demo-oai-public-net/) {
    print $0
    print "    public_net_access:"
    print "        external:"
    print "            name: oai-public-access"
  } else if ($0 ~/n2IpAddr:/) {
    n2IpAddrLine = $0
  } else if ($0 ~/n2Port:/) {
    n2PortLine = $0
  } else if ($0 ~/n3IpAddr: 192/) {
    n2IpAddrLine = $0
    n3IpAddrLine = $0
    gsub("n3IpAddr:", "n2IpAddr:", n2IpAddrLine)
    print n2IpAddrLine
    print n2PortLine
    gsub("70","72", n3IpAddrLine)
    print n3IpAddrLine
  } else if ($0 ~/defaultAs:/) {
    gsub("70","73", $0)
    print $0
  } else if ($0 ~/profileName/) {
    profileName = $0
    print $0
  } else if ($0 ~/enable: true/) {
    if (profileName ~/profile9/) {
      print $0
    } else {
      gsub("enable: true", "enable: false", $0)
      print $0
    }
  } else {
    print $0
  }
}
END{}
