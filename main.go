// +build linux

package main

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"golang.org/x/sys/unix"
	"log"
	_ "net/http/pprof"
	"runtime"
	"runtime/debug"
	"time"
)

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func printDevInfo(poll *epoll, ) {
	var ms runtime.MemStats
	var ms1 debug.GCStats

	for {
		runtime.ReadMemStats(&ms)
		debug.ReadGCStats(&ms1)
		fmt.Printf("%5d clients | HeapSys: %8d HeapInuse: %8d HeapReleased: %8d HeapObjects: %6d HeapAlloc: %8d HeapIdle: %8d NumGC: %8d PauseTotal: %d\n",
			len(poll.connections), ms.HeapSys, ms.HeapInuse, ms.HeapReleased, ms.HeapObjects, ms.HeapAlloc, ms.HeapIdle,ms1.NumGC,ms1.PauseTotal)
		time.Sleep(time.Second)
	}
}

func freeMemory() {
	for {
		time.Sleep(time.Second * 15)
		fmt.Println("run FreeOSMemory()")
		debug.FreeOSMemory()
	}
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())
	//go freeMemory()
	go pprofdump()

	tlsPort := flag.String("port", "8883", "")
	flag.Parse()

	cert, err := tls.LoadX509KeyPair("certs/server.pem", "certs/server.key")
	panicIfErr(err)

	config := &tls.Config{
		RootCAs:            x509.NewCertPool(),
		Certificates:       []tls.Certificate{cert},
		ClientAuth:         tls.RequestClientCert,
		ClientCAs:          nil,
		InsecureSkipVerify: true,
	}

	config.Rand = rand.Reader

	listener, err := tls.Listen("tcp", ":"+*tlsPort, config)
	panicIfErr(err)

	poll, _ := createPoll()
	go printDevInfo(poll)
	go poll.Wait()

	defer unix.Close(poll.fd)

	for {
		connection, err := listener.Accept()
		tlsconn, _ := connection.(*tls.Conn)
		err = tlsconn.Handshake()

		if err != nil {
			fmt.Println("tls error:", err)
			_ = tlsconn.Close()
			connection.Close()
			continue
		}
		if !verifyCert(tlsconn) {
			printConnState(tlsconn)
			fmt.Println("tls error: не правильный сертификат")
			_ = tlsconn.Close()
			connection.Close()
			continue
		}

		panicIfErr(err)
		poll.Add(connection)
	}
}

func verifyCert(conn *tls.Conn) bool {
	cert := conn.ConnectionState().PeerCertificates[0]
	issuer := cert.Issuer

	if !cert.IsCA || issuer.CommonName != "www.random.com" {
		return false
	}
	return true
}

func printConnState(conn *tls.Conn) {
	log.Print(">>>>>>>>>>>>>>>> State <<<<<<<<<<<<<<<<")
	state := conn.ConnectionState()
	log.Printf("Version: %x", state.Version)
	log.Printf("HandshakeComplete: %t", state.HandshakeComplete)
	log.Printf("DidResume: %t", state.DidResume)
	log.Printf("CipherSuite: %x", state.CipherSuite)
	log.Printf("NegotiatedProtocol: %s", state.NegotiatedProtocol)
	log.Printf("NegotiatedProtocolIsMutual: %t", state.NegotiatedProtocolIsMutual)

	log.Print("Certificate chain:")
	for i, cert := range state.PeerCertificates {

		subject := cert.Subject
		issuer := cert.Issuer
		log.Printf("Serial Number: %s", cert.SerialNumber)
		log.Printf("Signature Algorithm: %s", cert.SignatureAlgorithm)
		log.Printf("IsCA: %t", cert.IsCA)
		//log.Printf("Signature: %s", cert.Signature)

		log.Printf(" %d s:/C=%v/ST=%v/L=%v/O=%v/OU=%v/CN=%s", i, subject.Country, subject.Province, subject.Locality, subject.Organization, subject.OrganizationalUnit, subject.CommonName)
		log.Printf("   i:/C=%v/ST=%v/L=%v/O=%v/OU=%v/CN=%s", issuer.Country, issuer.Province, issuer.Locality, issuer.Organization, issuer.OrganizationalUnit, issuer.CommonName)
	}
	log.Print(">>>>>>>>>>>>>>>> State End <<<<<<<<<<<<<<<<")
}
