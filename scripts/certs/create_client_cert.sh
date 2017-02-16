#!/usr/bin/env bash

NAME=$1

if [ -z $NAME ]; then
	echo set username
	exit 1;
fi

UNAME=`getent passwd $NAME | cut -d: -f1`
USID=`getent passwd $NAME | cut -d: -f3`
GRID=`getent passwd $NAME | cut -d: -f4`

echo make certificate for user $UNAME, id: $USID, groupid: $GRID

openssl genrsa -out key_$UNAME.pem 4096
openssl req -subj "/CN=$USID:$GRID" -new -key key_$UNAME.pem -out client.csr

echo extendedKeyUsage = clientAuth > extfile_client.cnf


openssl x509 -req -days 365 -sha256 -in client.csr -CA ca.pem -CAkey ca-key.pem \
	  -CAcreateserial -out cert_$UNAME.pem -extfile extfile_client.cnf \
	  -passin file:passphrase.txt

rm -v client.csr ca.srl extfile_client.cnf


chmod -v 0400 key_$UNAME.pem
chmod -v 0444 cert_$UNAME.pem
chown $UNAME: key_$UNAME.pem
chown $UNAME: cert_$UNAME.pem


