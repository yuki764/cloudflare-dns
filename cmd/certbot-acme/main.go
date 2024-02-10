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
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	client := &dns.Client{Token: os.Getenv("TOKEN"), ZoneId: os.Getenv("ZONE_ID")}

	comment := os.Getenv("COMMENT")
	validationToken := os.Getenv("CERTBOT_VALIDATION")

	r, err := client.GetRecord(dns.Record{
		Name:    "_acme-challenge." + os.Getenv("CERTBOT_DOMAIN"),
		Type:    "TXT",
		Comment: comment,
	})
	if err != nil {
		panic(err)
	}

	slog.Debug("debug", "debug", r)

	if err := client.PatchRecord(dns.Record{
		Name:    r.Name,
		Type:    r.Type,
		Content: validationToken, // replace validation record only
		Comment: r.Comment,
		Id:      r.Id,
	}); err != nil {
		panic(err)
	}

	// confirm TXT record updated
	checkNum := 10
	for i := range checkNum + 1 {
		time.Sleep(10 * time.Second)

		slog.Info(fmt.Sprintf("checking if validation record is updated (%d of %d)", i+1, checkNum))
		vals, err := net.LookupTXT(r.Name)
		if err != nil {
			panic(err)
		}
		for _, v := range vals {
			if v == validationToken {
				slog.Info("confirm validation record is updated")
				return
			}
		}
	}
}
