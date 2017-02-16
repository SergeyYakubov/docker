#!/usr/bin/env bash

HOST=localhost
ORG=yourorg

openssl genrsa -aes256 -passout file:passphrase.txt -out ca-key.pem 4096
openssl req -new -x509 -days 365 -key ca-key.pem -sha256 -out ca.pem -subj "/CN=$HOST/O=$ORG/C=DE" -passin file:passphrase.txt



chmod -v 0400 ca-key.pem
chmod -v 0444 ca.pem
