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

	//"net/http/httputil"

	"github.com/gorilla/mux"
	"golang.org/x/net/http2"
)

var ()

const ()

func eventsMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		fmt.Println("PeerCertificates")
		for _, cert := range r.TLS.PeerCertificates {
			fmt.Printf("  Subject %s\n", cert.Subject)
			fmt.Printf("  Issuer Name: %s\n", cert.Issuer)
			fmt.Printf("  Expiry: %s \n", cert.NotAfter.Format("2006-January-02"))
			fmt.Printf("  Issuer Common Name: %s \n", cert.Issuer.CommonName)
			fmt.Printf("  IsCA: %t \n", cert.IsCA)

			hasher := sha256.New()
			hasher.Write(cert.Raw)
			clientCertificateHash := base64.StdEncoding.EncodeToString(hasher.Sum(nil))

			fmt.Printf("  Certificate hash %s\n", clientCertificateHash)
			fmt.Println()

		}
		fmt.Println("VerifiedChains")
		for _, cert := range r.TLS.VerifiedChains {
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

		h.ServeHTTP(w, r)
	})
}

func gethandler(w http.ResponseWriter, r *http.Request) {

	fmt.Fprint(w, "ok")
}

func posthandler(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	log.Printf("Data val [%s]", string(body))

	fmt.Fprint(w, "ok")
}

func main() {

	router := mux.NewRouter()
	router.Methods(http.MethodGet).Path("/").HandlerFunc(gethandler)
	router.Methods(http.MethodPost).Path("/").HandlerFunc(posthandler)

	var err error
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

	tlsConfig := &tls.Config{
		ClientAuth: tls.RequireAndVerifyClientCert,
		ClientCAs:  caCertPool,
	}

	server := &http.Server{
		Addr:      ":8081",
		Handler:   eventsMiddleware(router),
		TLSConfig: tlsConfig,
	}
	http2.ConfigureServer(server, &http2.Server{})
	fmt.Println("Starting Server..")
	err = server.ListenAndServeTLS("certs/server.crt", "certs/server.key")
	fmt.Printf("Unable to start Server %v", err)

}
