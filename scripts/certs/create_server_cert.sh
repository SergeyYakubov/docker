#!/usr/bin/env bash

HOST=$1

if [ -z $HOST ]; then
 echo set hostname
 exit 1;
fi


openssl genrsa -out server-key_$HOST.pem 4096
openssl req -subj "/CN=$HOST" -sha256 -new -key server-key_$HOST.pem -out server.csr

echo subjectAltName = DNS:$HOST,IP:127.0.0.1 > extfile.cnf

openssl x509 -req -days 365 -sha256 -in server.csr -CA ca.pem -CAkey ca-key.pem \
  -CAcreateserial -out server-cert_$HOST.pem -extfile extfile.cnf -passin file:passphrase.txt

rm server.csr ca.srl extfile.cnf

chmod -v 0400 server-key_$HOST.pem
chmod -v 0444 server-cert_$HOST.pem
