#!/bin/bash

host=localhost:32888
ip=192.168.1.254
#ip=$(curl ifconfig.me)

read -r -d '' headers <<EOF
Secret: TEST
IP: $ip
Name: localserver
EOF

curl -X GET -H "$headers" $host

