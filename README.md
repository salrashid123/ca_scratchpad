## Create Root CA Key and cert

This repo is just a simple template to create a self-issued CA and Subordinate CA to issue TLS certificates

```
$ tree
.
├── ca            // CA details stored here
├── certs         // issued certs stored here
├── client.conf   // config for the client certs 
├── crl           // CRLs are stored here
├── LICENSE
├── README.md
├── root-ca.conf  // config for the root ca
├── server.conf   // config for the server certs 
└── tls-ca.conf   // config for the subordinate ca 
```


### Create Root CA

Create a root CA and subordinate

```bash
mkdir -p ca/root-ca/private ca/root-ca/db crl certs
chmod 700 ca/root-ca/private
cp /dev/null ca/root-ca/db/root-ca.db
cp /dev/null ca/root-ca/db/root-ca.db.attr

echo 01 > ca/root-ca/db/root-ca.crt.srl
echo 01 > ca/root-ca/db/root-ca.crl.srl


openssl req -new     -config root-ca.conf     -out ca/root-ca.csr     -keyout ca/root-ca/private/root-ca.key
 (pick any password)
openssl ca -selfsign     -config root-ca.conf     -in ca/root-ca.csr     -out ca/root-ca.crt     -extensions root_ca_ext
openssl x509 -in ca/root-ca.crt -text -noout
```


### Gen CRL

Optionally create a CRL

```bash
mkdir crl/
openssl ca -gencrl     -config root-ca.conf     -out crl/root-ca.crl
openssl crl -in crl/root-ca.crl -noout -text
```

### Create Subordinate CA for TLS Signing

This is the CA that will issue the client and server certs

```
mkdir -p ca/tls-ca/private ca/tls-ca/db crl certs
chmod 700 ca/tls-ca/private

cp /dev/null ca/tls-ca/db/tls-ca.db
cp /dev/null ca/tls-ca/db/tls-ca.db.attr
echo 01 > ca/tls-ca/db/tls-ca.crt.srl
echo 01 > ca/tls-ca/db/tls-ca.crl.srl


openssl req -new \
    -config tls-ca.conf \
    -out ca/tls-ca.csr \
    -keyout ca/tls-ca/private/tls-ca.key


openssl ca \
    -config root-ca.conf \
    -in ca/tls-ca.csr \
    -out ca/tls-ca.crt \
    -extensions signing_ca_ext


openssl ca -gencrl \
    -config tls-ca.conf \
    -out crl/tls-ca.crl
```

- Combine the Root and subordinate CA into a chain:

```bash
cat ca/tls-ca.crt ca/root-ca.crt >  ca/tls-ca-chain.pem
```

### Generate Server Cert (TLS)

Set env vars for the server cert

NAME: filename to use for the cert/key/csr
SAN:  SAN Value

edit the CN value below (eg to the same SAN Value)

```bash
export NAME=server
export SAN=DNS:server.domain.com
openssl req -new     -config server.conf     -out certs/$NAME.csr     -keyout certs/$NAME.key -subj "/C=US/O=Google/OU=Enterprise/CN=server.domain.com"

openssl ca \
    -config tls-ca.conf \
    -in certs/$NAME.csr \
    -out certs/$NAME.crt \
    -extensions server_ext
```


## Generate Client Certificate

Generate client certificate

```bash
export NAME=tokenclient

openssl req -new \
    -config client.conf \
    -out certs/$NAME.csr \
    -keyout certs/$NAME.key \
    -subj "/C=US/O=Google/OU=Enterprise/CN=user@domain.com"

openssl ca \
    -config tls-ca.conf \
    -in certs/$NAME.csr \
    -out certs/$NAME.crt \
    -policy extern_pol 
```

### Revoke a certificate

If you want to revoke a certificate `certs/$NAME.crt`, run

```
openssl ca -config tls-ca.conf   -revoke certs/$NAME.crt
```
and then regenerate the CRL

