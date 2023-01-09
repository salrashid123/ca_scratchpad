package main

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

var ()

func main() {

	caCertPool := x509.NewCertPool()
	caCert, err := ioutil.ReadFile("certs/root-ca-1.crt")
	if err != nil {
		log.Println(err)
		return
	}
	caCertPool.AppendCertsFromPEM(caCert)

	caCert, err = ioutil.ReadFile("certs/root-ca-2.crt")
	if err != nil {
		log.Println(err)
		return
	}
	caCertPool.AppendCertsFromPEM(caCert)

	caCert, err = ioutil.ReadFile("certs/tls-ca-1.crt")
	if err != nil {
		log.Println(err)
		return
	}
	caCertPool.AppendCertsFromPEM(caCert)

	caCert, err = ioutil.ReadFile("certs/tls-ca-cross.crt")
	if err != nil {
		log.Println(err)
		return
	}
	caCertPool.AppendCertsFromPEM(caCert)

	clientCerts, err := tls.LoadX509KeyPair(
		"certs/client.crt",
		"certs/client.key",
	)

	tlsConfig := &tls.Config{
		ServerName:   "server.domain.com",
		Certificates: []tls.Certificate{clientCerts},
		RootCAs:      caCertPool,
		VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {

			fmt.Println("VerifiedChains")
			for _, cert := range verifiedChains {
				for _, c := range cert {
					fmt.Printf("  Subject %s\n", c.Subject)
					fmt.Printf("  Issuer Name: %s\n", c.Issuer)
					fmt.Printf("  Expiry: %s \n", c.NotAfter.Format("2006-January-02"))
					fmt.Printf("  Issuer Common Name: %s \n", c.Issuer.CommonName)
					fmt.Printf("  IsCA: %t \n", c.IsCA)
					h := sha256.New()
					h.Write(c.Raw)
					clientCertificateHash := base64.StdEncoding.EncodeToString(h.Sum(nil))

					fmt.Printf("  Certificate hash %s\n", clientCertificateHash)
					fmt.Println()
				}
			}
			return nil

		},
	}

	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get("https://server.domain.com:8081")
	if err != nil {
		log.Println(err)
		return
	}

	htmlData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("%v\n", resp.Status)
	fmt.Printf(string(htmlData))

}
