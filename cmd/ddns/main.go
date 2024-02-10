package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"go.neigepluie.net/cloudflare-dns/pkg/cloudflare/dns"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	client := &dns.Client{Token: os.Getenv("TOKEN"), ZoneId: os.Getenv("ZONE_ID")}

	// set updating interval
	interval, err := time.ParseDuration(os.Getenv("INTERVAL"))
	if err != nil {
		panic(err)
	}

	// list records with specific comment
	commentPrefix := os.Getenv("COMMENT_PREFIX")
	if commentPrefix == "" {
		slog.Error("environment variable COMMENT_PREFIX must be specified")
		panic(nil)
	}
	records, err := client.ListRecordsHasCommentPrefix(commentPrefix)
	if err != nil {
		panic(err)
	}
	slog.Debug("debug", "debug", records)

	for {
		if len(records) > 0 {
			myAddr := make(map[string]string)

			// get external IPv4 Address
			myIpv4Addr, err := getExternalAddr(4)
			if err != nil {
				slog.Warn("failed to get IPv4 address", "error", err)
			} else {
				myAddr["A"] = myIpv4Addr
				slog.Info("My IPv4 address is " + myAddr["A"])
			}

			// get external IPv6 Address
			myIpv6Addr, err := getExternalAddr(6)
			if err != nil {
				slog.Warn("failed to get IPv6 address", "error", err)
			} else {
				myAddr["AAAA"] = myIpv6Addr
				slog.Info("My IPv6 address is " + myAddr["AAAA"])
			}

			updateTime := time.Now().Format(time.RFC3339)

			// update records
			for _, r := range records {
				addr, ok := myAddr[r.Type]
				if ok {
					if err := client.PatchRecord(dns.Record{
						Name:    r.Name,
						Type:    r.Type,
						Content: addr, // update current my address
						Id:      r.Id,
						Comment: commentPrefix + "updated/" + updateTime,
					}); err != nil {
						slog.Warn("failed to update record", "error", err)
					}
				} else {
					slog.Error(r.Type + " record type does not supported in this environment")
					panic(nil)
				}
			}
		} else {
			slog.Info("nothing to update")
		}

		time.Sleep(interval)
	}
}

// using https://ipinfo.io
func getExternalAddr(ipVersion int) (string, error) {
	ipinfoUrl := ""
	if ipVersion == 4 {
		ipinfoUrl = "https://ipinfo.io"
	} else if ipVersion == 6 {
		ipinfoUrl = "https://v6.ipinfo.io"
	} else {
		return "", errors.New("unknown Internet Protocol stack.")
	}

	resp, err := http.Get(ipinfoUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var ipinfoResp struct {
		Ip string `json:"ip"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ipinfoResp); err != nil {
		return "", err
	}

	return ipinfoResp.Ip, err
}
