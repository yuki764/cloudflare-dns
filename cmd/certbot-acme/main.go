package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	"go.neigepluie.net/cloudflare-dns/pkg/cloudflare/dns"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	client := &dns.Client{Token: os.Getenv("TOKEN"), ZoneId: os.Getenv("ZONE_ID")}

	recordName := "_acme-challenge." + os.Getenv("CERTBOT_DOMAIN")

	r, err := client.GetRecordId(dns.Record{
		Name: recordName,
		Type: "TXT",
	})
	if err != nil {
		panic(err)
	}

	slog.Info("debug", "debug", r)

	if err := client.PatchRecord(dns.Record{
		Name:    r.Name,
		Type:    r.Type,
		Content: os.Getenv("CERTBOT_VALIDATION"), // replace validation record only
		Id:      r.Id,
	}); err != nil {
		panic(err)
	}

	// confirm TXT record updated
	checkNum := 5
	for i := range checkNum + 1 {
		time.Sleep(30 * time.Second)

		slog.Info(fmt.Sprintf("checking if validation record is updated (%d of %d)", i+1, checkNum))
		vals, err := net.LookupTXT(recordName)
		if err != nil {
			panic(err)
		}
		if vals[0] == os.Getenv("CERTBOT_VALIDATION") {
			slog.Info("confirm validation record is updated")
			break
		}
	}
}
