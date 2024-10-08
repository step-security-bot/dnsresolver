// Package dnsrecordcache provides a simple in-memory cache
package dnsrecordcache

import (
	"dnsresolver/dnsrecords"
	"fmt"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// CacheRecord holds the data for the cache records
type CacheRecord struct {
	DNSRecord dnsrecords.DNSRecord `json:"dns_record"`
	Expiry    time.Time            `json:"expiry,omitempty"`
	Timestamp time.Time            `json:"timestamp,omitempty"`
	LastQuery time.Time            `json:"last_query,omitempty"`
}

// Add a new record to the cache
func Add(cacheRecordsData []CacheRecord, record *dns.RR) []CacheRecord {
	var value string

	switch r := (*record).(type) {
	case *dns.A:
		value = r.A.String()
	case *dns.AAAA:
		value = r.AAAA.String()
	case *dns.CNAME:
		value = r.Target
	case *dns.MX:
		value = fmt.Sprintf("%d %s", r.Preference, r.Mx)
	case *dns.NS:
		value = r.Ns
	case *dns.SOA:
		value = fmt.Sprintf("%s %s %d %d %d %d %d", r.Ns, r.Mbox, r.Serial, r.Refresh, r.Retry, r.Expire, r.Minttl)
	case *dns.TXT:
		value = strings.Join(r.Txt, " ")
	default:
		value = (*record).String()
	}

	cacheRecord := CacheRecord{
		DNSRecord: dnsrecords.DNSRecord{
			Name:  (*record).Header().Name,
			Type:  dns.TypeToString[(*record).Header().Rrtype],
			Value: value,
			TTL:   (*record).Header().Ttl,
		},
		Expiry:    time.Now().Add(time.Duration((*record).Header().Ttl) * time.Second),
		Timestamp: time.Now(),
	}

	// Check if the record already exists in the cache
	recordIndex := -1
	for i, existingRecord := range cacheRecordsData {
		if existingRecord.DNSRecord.Name == cacheRecord.DNSRecord.Name &&
			existingRecord.DNSRecord.Type == cacheRecord.DNSRecord.Type &&
			existingRecord.DNSRecord.Value == cacheRecord.DNSRecord.Value {
			recordIndex = i
			break
		}
	}

	// If the record exists in the cache, update its TTL, expiry, and last query, otherwise add it
	if recordIndex != -1 {
		cacheRecordsData[recordIndex].DNSRecord.TTL = cacheRecord.DNSRecord.TTL
		cacheRecordsData[recordIndex].Expiry = cacheRecord.Expiry
		cacheRecordsData[recordIndex].LastQuery = time.Now()
	} else {
		cacheRecord.LastQuery = time.Now()
		cacheRecordsData = append(cacheRecordsData, cacheRecord)
	}

	return cacheRecordsData
}

// List all records in the cache
func List(cacheRecordsData []CacheRecord) {
	fmt.Println("Cache Records:")
	for i, record := range cacheRecordsData {
		fmt.Printf("%d. %s %s %s %d\n", i+1, record.DNSRecord.Name, record.DNSRecord.Type, record.DNSRecord.Value, record.DNSRecord.TTL)
	}
}

// Remove a record from the cache
func Remove(fullCommand []string, cacheRecordsData []CacheRecord) []CacheRecord {
	if len(fullCommand) > 1 && fullCommand[1] == "?" {
		fmt.Println("Enter the cache record in the format: <Name>")
		fmt.Println("Example: example.com")
		return nil
	}

	if len(fullCommand) < 2 {
		fmt.Println("Please specify at least the record name.")
		return nil
	}

	name := fullCommand[1]

	matchingRecords := []CacheRecord{}
	for _, record := range cacheRecordsData {
		if record.DNSRecord.Name == name {
			matchingRecords = append(matchingRecords, record)
		}
	}

	if len(matchingRecords) == 0 {
		fmt.Println("No records found with the name:", name)
		return nil
	}

	if len(fullCommand) == 2 {
		if len(matchingRecords) == 1 {
			for i, r := range cacheRecordsData {
				if r == matchingRecords[0] {
					cacheRecordsData = append(cacheRecordsData[:i], cacheRecordsData[i+1:]...)
					removedRecToPrint := fmt.Sprintf("%s %s %s %d", matchingRecords[0].DNSRecord.Name, matchingRecords[0].DNSRecord.Type, matchingRecords[0].DNSRecord.Value, matchingRecords[0].DNSRecord.TTL)
					fmt.Println("Removed:", removedRecToPrint)
					return cacheRecordsData
				}
			}
			return nil
		}
		fmt.Println("Multiple records found with the name:", name)
		for i, record := range matchingRecords {
			fmt.Printf("%d. %v\n", i+1, record)
		}
		return nil
	}

	return cacheRecordsData
}
