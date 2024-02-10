package dns

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

type Record struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
	Comment string `json:"comment,omitempty"`
	Id      string `json:"id,omitempty"`
}

type Client struct {
	Token  string
	ZoneId string
}

func (c Client) GetRecord(r Record) (*Record, error) {
	cloudflareEndpoint := "https://api.cloudflare.com/client/v4/zones/" + c.ZoneId + "/dns_records"

	// get record ID from name and record type
	req, err := http.NewRequest("GET", cloudflareEndpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+c.Token)

	q := req.URL.Query()
	q.Add("name", r.Name)
	q.Add("type", r.Type)
	if r.Comment != "" {
		q.Add("comment", r.Comment)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var listRecords struct {
		Result []Record `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&listRecords); err != nil {
		return nil, err
	}
	if len(listRecords.Result) != 1 {
		slog.Error("query result is wrong. the record does not exist or be duplicated.")
		return nil, err
	}

	return &listRecords.Result[0], nil
}

func (c Client) ListRecordsHasCommentPrefix(commentPrefix string) ([]Record, error) {
	cloudflareEndpoint := "https://api.cloudflare.com/client/v4/zones/" + c.ZoneId + "/dns_records"

	// get record ID from name and record type
	req, err := http.NewRequest("GET", cloudflareEndpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+c.Token)

	q := req.URL.Query()
	q.Add("comment.startswith", commentPrefix)
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var listRecords struct {
		Result []Record `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&listRecords); err != nil {
		return nil, err
	}

	return listRecords.Result, nil
}

func (c Client) PatchRecord(r Record) error {
	cloudflareEndpoint := "https://api.cloudflare.com/client/v4/zones/" + c.ZoneId + "/dns_records/" + r.Id

	patchData, err := json.Marshal(Record{r.Name, r.Type, r.Content, r.Comment, ""})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH", cloudflareEndpoint, bytes.NewReader(patchData))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+c.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var result interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return err
		}
		slog.Info("patch request succeeded", "result", result)
	} else {
		return errors.New("patch request failed")
	}

	return nil
}
