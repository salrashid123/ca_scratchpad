## Create Root CA Key and cert

This repo is just a simple template to create a self-issued CA and Subordinate CA to issue TLS certificates.

The `cross_sign/` folder contains instruction to create two CAs and Subordinate CAs where one is cross signed

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

This will generate a root CA, crls, subordinate CA and then using the subordinate, any number of client and server certs.

If you just want leaf certs issued directly from the root, see the section at the end 

For starters, we'll just crate a single root CA:

### Single Level CA

The following will issue certificates directly from the root


```bash
mkdir -p ca/root-ca/private ca/root-ca/db crl certs
chmod 700 ca/root-ca/private
cp /dev/null ca/root-ca/db/root-ca.db
cp /dev/null ca/root-ca/db/root-ca.db.attr

echo 01 > ca/root-ca/db/root-ca.crt.srl
echo 01 > ca/root-ca/db/root-ca.crl.srl

# Pick signature algo (either do A,B or C)

# A) Signature Algorithm: sha256WithRSAEncryption
    openssl genpkey -algorithm rsa -pkeyopt rsa_keygen_bits:2048 \
      -pkeyopt rsa_keygen_pubexp:65537 -out ca/root-ca/private/root-ca.key

# B) Signature Algorithm: rsassaPss
    openssl genpkey -algorithm rsa-pss -pkeyopt rsa_keygen_bits:2048 \
      -pkeyopt rsa_keygen_pubexp:65537 -out ca/root-ca/private/root-ca.key

# C) Signature Algorithm: ecdsa-with-SHA256
    openssl genpkey -algorithm ec -pkeyopt  ec_paramgen_curve:P-256 \
      -pkeyopt ec_param_enc:named_curve -pkeyopt ec_paramgen_curve:secp384r1 \
      -out ca/root-ca/private/root-ca.key
   
openssl req -new  -config single-root-ca.conf  -key ca/root-ca/private/root-ca.key \
   -out ca/root-ca.csr  

openssl ca -selfsign     -config single-root-ca.conf  \
   -in ca/root-ca.csr     -out ca/root-ca.crt  \
   -extensions root_ca_ext

```

- Issue server

```bash
export NAME=server

openssl genpkey -algorithm rsa -pkeyopt rsa_keygen_bits:2048 \
      -pkeyopt rsa_keygen_pubexp:65537 -out certs/$NAME.key

openssl req -new     -config server.conf \
  -out certs/$NAME.csr  \
  -key certs/$NAME.key \
  -subj "/C=US/O=Google/OU=Enterprise/CN=server.domain.com" 

## to specify a SAN, edit single-root-ca.conf and modify [ alt_names ]

openssl ca \
    -config single-root-ca.conf \
    -in certs/$NAME.csr \
    -out certs/$NAME.crt  \
    -extensions server_ext
```

- Issue client


```bash
export NAME=user10

openssl genpkey -algorithm rsa -pkeyopt rsa_keygen_bits:2048 \
      -pkeyopt rsa_keygen_pubexp:65537 -out certs/$NAME.key

openssl req -new \
    -config client.conf \
    -out certs/$NAME.csr \
    -key certs/$NAME.key \
    -subj "/L=US/O=Google/OU=Enterprise/CN=user10.domain.com"

## to specify a SAN, edit single-root0ca.conf and modify [ alt_names ]

openssl ca \
    -config single-root-ca.conf \
    -in certs/$NAME.csr \
    -out certs/$NAME.crt \
    -policy extern_pol \
    -extensions client_ext
```


note, for ECC keys, its something like this:

```bash
openssl genpkey -algorithm ec -pkeyopt  ec_paramgen_curve:P-256 \
      -pkeyopt ec_param_enc:named_curve \
      -out certs/$NAME.key

openssl req -new     -config server.conf \
  -out certs/$NAME.csr   \
  -key certs/$NAME.key \
  -subj "/C=US/O=Google/OU=Enterprise/CN=server.domain.com"

openssl ca \
    -config single-root-ca.conf \
    -in certs/$NAME.csr \
    -out certs/$NAME.crt \
    -extensions server_ext
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

# Pick signature algo (either do A,B or C)

# A) Signature Algorithm: sha256WithRSAEncryption
    openssl genpkey -algorithm rsa -pkeyopt rsa_keygen_bits:2048 \
      -pkeyopt rsa_keygen_pubexp:65537 -out ca/root-ca/private/root-ca.key

# B) Signature Algorithm: rsassaPss
    openssl genpkey -algorithm rsa-pss -pkeyopt rsa_keygen_bits:2048 \
      -pkeyopt rsa_keygen_pubexp:65537 -out ca/root-ca/private/root-ca.key

# C) Signature Algorithm: ecdsa-with-SHA256
    openssl genpkey -algorithm ec -pkeyopt  ec_paramgen_curve:P-256 \
      -pkeyopt ec_param_enc:named_curve -pkeyopt ec_paramgen_curve:secp384r1 \
      -out ca/root-ca/private/root-ca.key
   
openssl req -new  -config root-ca.conf  -key ca/root-ca/private/root-ca.key \
   -out ca/root-ca.csr  

openssl ca -selfsign     -config root-ca.conf  \
   -in ca/root-ca.csr     -out ca/root-ca.crt  \
   -extensions root_ca_ext

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

```bash
mkdir -p ca/tls-ca/private ca/tls-ca/db crl certs
chmod 700 ca/tls-ca/private

cp /dev/null ca/tls-ca/db/tls-ca.db
cp /dev/null ca/tls-ca/db/tls-ca.db.attr
echo 01 > ca/tls-ca/db/tls-ca.crt.srl
echo 01 > ca/tls-ca/db/tls-ca.crl.srl

# Pick signature algo (either do A,B or C)

# A) Signature Algorithm: sha256WithRSAEncryption
    openssl genpkey -algorithm rsa -pkeyopt rsa_keygen_bits:2048 \
      -pkeyopt rsa_keygen_pubexp:65537 -out ca/tls-ca/private/tls-ca.key

# B) Signature Algorithm: rsassaPss
    openssl genpkey -algorithm rsa-pss -pkeyopt rsa_keygen_bits:2048 \
      -pkeyopt rsa_keygen_pubexp:65537 -out ca/tls-ca/private/tls-ca.key

# C) Signature Algorithm: ecdsa-with-SHA256
    openssl genpkey -algorithm ec -pkeyopt  ec_paramgen_curve:P-256 \
      -pkeyopt ec_param_enc:named_curve -pkeyopt ec_paramgen_curve:secp384r1 \
      -out ca/tls-ca/private/tls-ca.key
   

openssl req -new  -config tls-ca.conf -key ca/tls-ca/private/tls-ca.key \
   -out ca/tls-ca.csr

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

openssl genpkey -algorithm rsa -pkeyopt rsa_keygen_bits:2048 \
      -pkeyopt rsa_keygen_pubexp:65537 -out certs/$NAME.key

openssl req -new     -config server.conf \
  -out certs/$NAME.csr   \
  -key certs/$NAME.key \
  -subj "/C=US/O=Google/OU=Enterprise/CN=server.domain.com"

## to specify a SAN, edit tls-caa.conf and modify [ alt_names ]

openssl ca \
    -config tls-ca.conf \
    -in certs/$NAME.csr \
    -out certs/$NAME.crt \
    -extensions server_ext
```


## Generate Client Certificate

Generate client certificate

```bash
export NAME=user

openssl genpkey -algorithm rsa -pkeyopt rsa_keygen_bits:2048 \
      -pkeyopt rsa_keygen_pubexp:65537 -out certs/$NAME.key


openssl req -new \
    -config client.conf \
    -out certs/$NAME.csr \
    -key certs/$NAME.key \
    -subj "/C=US/O=Google/OU=Enterprise/CN=user@domain.com"

## to specify a SAN, edit tls-caa.conf and modify [ alt_names ]

openssl ca \
    -config tls-ca.conf \
    -in certs/$NAME.csr \
    -out certs/$NAME.crt \
    -policy extern_pol \
    -extensions client_ext
```

### Revoke a certificate

If you want to revoke a certificate `certs/$NAME.crt`, run

```bash
openssl ca -config tls-ca.conf   -revoke certs/$NAME.crt
```
and then regenerate the CRL

you'll end up with

```bash
$ tree
.
├── ca
│   ├── root-ca
│   │   ├── 01.pem
│   │   ├── 02.pem
│   │   ├── db
│   │   │   ├── root-ca.crl.srl
│   │   │   ├── root-ca.crl.srl.old
│   │   │   ├── root-ca.crt.srl
│   │   │   ├── root-ca.crt.srl.old
│   │   │   ├── root-ca.db
│   │   │   ├── root-ca.db.attr
│   │   │   ├── root-ca.db.attr.old
│   │   │   └── root-ca.db.old
│   │   └── private
│   │       └── root-ca.key                // root CA key
│   ├── root-ca.crt                        // root CA cert
│   ├── root-ca.csr
│   ├── tls-ca
│   │   ├── 01.pem
│   │   ├── 02.pem
│   │   ├── db
│   │   │   ├── tls-ca.crl.srl
│   │   │   ├── tls-ca.crl.srl.old
│   │   │   ├── tls-ca.crt.srl
│   │   │   ├── tls-ca.crt.srl.old
│   │   │   ├── tls-ca.db
│   │   │   ├── tls-ca.db.attr
│   │   │   ├── tls-ca.db.attr.old
│   │   │   └── tls-ca.db.old
│   │   └── private
│   │       └── tls-ca.key                // subordinate CA key
│   ├── tls-ca-chain.pem
│   ├── tls-ca.crt                        // subordinate CA cert
│   └── tls-ca.csr
├── certs
│   ├── server.crt                        // issued server cert by tls-ca
│   ├── server.csr 
│   ├── server.key                        // issued server key by tls-ca
│   ├── tokenclient.crt                   // client cert
│   ├── tokenclient.csr
│   └── tokenclient.key                   // client key
├── client.conf
├── crl                                   // CRL file
│   ├── root-ca.crl
│   └── tls-ca.crl
├── LICENSE
├── README.md
├── root-ca.conf
├── server.conf
└── tls-ca.conf
```

### TPM based private key

If you have openssl and want to issue a cert on the TPM, 

also see [sa_pki.sh](https://github.com/tpm2-software/tpm2-openssl/blob/master/test/rsa_pki/rsa_pki.sh)

using openssl3 [tpm2-openssl](https://github.com/tpm2-software/tpm2-openssl) installed:

```bash
openssl version
   OpenSSL 3.0.9 30 May 2023 (Library: OpenSSL 3.0.9 30 May 2023)

export NAME=tpms

openssl genpkey -provider tpm2 -algorithm RSA -pkeyopt rsa_keygen_bits:2048 \
      -pkeyopt rsa_keygen_pubexp:65537 -out certs/$NAME.key

openssl req -new  -provider tpm2 -provider default \
      -config server.conf   -out certs/$NAME.csr \
          -key certs/$NAME.key   -subj "/C=US/O=Google/OU=Enterprise/CN=server.domain.com"

openssl ca \
    -config single-root-ca.conf  \
    -in certs/$NAME.csr \
    -out certs/$NAME.crt \
    -extensions server_ext

openssl  x509 -pubkey -noout -in certs/$NAME.crt

openssl rsa  -provider tpm2 -provider default -in certs/$NAME.key -pubout 
```
