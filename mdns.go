package main

import (
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/miekg/dns"
)

func startMDNSServer(port string) {
	portInt, _ := strconv.Atoi(port)

	// Set up the multicast address for mDNS
	addr := &net.UDPAddr{IP: net.ParseIP("224.0.0.251"), Port: portInt}

	// Create a UDP connection to listen on multicast address
	conn, err := net.ListenMulticastUDP("udp4", nil, addr)
	if err != nil {
		log.Fatalf("Error setting up mDNS server: %v", err)
	}

	// Set reuse address option to allow multiple listeners on the same address
	if err := conn.SetReadBuffer(65535); err != nil {
		log.Fatalf("Failed to set read buffer size: %v", err)
	}

	server := &dns.Server{
		PacketConn: conn,
	}

	dns.HandleFunc("local.", handleMDNSRequest)

	log.Printf("Starting mDNS server on %s\n", addr)
	if err := server.ActivateAndServe(); err != nil {
		log.Fatalf("Error starting mDNS server: %v", err)
	}
}

// This is JUST a test, it will always return the same IP :P
func handleMDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	if len(r.Question) == 0 {
		return
	}

	q := r.Question[0]
	log.Printf("Received mDNS query: %s %s\n", q.Name, dns.TypeToString[q.Qtype])

	// Check if the request is for the .local domain
	if q.Qclass != dns.ClassINET || !dns.IsSubDomain("local.", q.Name) {
		log.Printf("Not an mDNS query, ignoring: %s\n", q.Name)
		return
	}

	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch q.Qtype {
	case dns.TypeA:
		// Return IPv4 address for A query
		ipv4 := net.ParseIP("127.0.0.1")
		if ipv4 == nil {
			log.Printf("Invalid IPv4 address provided\n")
			return
		}
		m.Answer = append(m.Answer, &dns.A{
			Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 120},
			A:   ipv4,
		})
	case dns.TypeAAAA:
		// Return IPv6 address for AAAA query
		ipv6 := net.ParseIP("::1")
		if ipv6 == nil {
			log.Printf("Invalid IPv6 address provided\n")
			return
		}
		m.Answer = append(m.Answer, &dns.AAAA{
			Hdr:  dns.RR_Header{Name: q.Name, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: 120},
			AAAA: ipv6,
		})
	default:
		log.Printf("Unsupported mDNS query type: %d\n", q.Qtype)
		return
	}

	// Write the response to the multicast address
	if err := w.WriteMsg(m); err != nil {
		log.Printf("Failed to write mDNS response: %v\n", err)
	}
}

func isMDNSQuery(name string) bool {
	// Check if the query is for the .local domain
	if strings.HasSuffix(name, ".local.") {
		return true
	}

	// Split the query name by dots
	parts := strings.Split(name, ".")

	// Check if the query has at least four parts (minimum for an mDNS query)
	if len(parts) < 3 {
		return false
	}

	return false
}
