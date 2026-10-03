package server

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/octoplorer/octopulse/internal/config"
	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/store"
)

func TestMonitorCollectionsAcrossPersistence(t *testing.T) {
	s, ts, c := testServer(t)
	csrf := bootstrap(t, c, ts.URL)
	status, body := request(t, c, "POST", ts.URL+"/api/v1/pages", csrf, domain.Page{
		Name: "Invalid collection", Slug: "invalid",
		Draft: domain.PageConfig{Title: "Status", BrandColor: "#008877", ColorScheme: "system", Links: []domain.Link{}},
	})
	if status != 422 || !bytes.Contains(body, []byte("body.draft.groups")) {
		t.Fatalf("required null collection accepted: %d %s", status, body)
	}
	m := createTestMonitor(t, c, ts.URL, csrf)
	if m.Tags == nil || m.NotificationChannelIDs == nil || m.HTTP.Headers == nil {
		t.Fatalf("creation returned null collections: %+v", m)
	}
	// Existing records may still contain null arrays. Reading them must not
	// require a data migration or frontend normalization.
	m.Tags = nil
	m.NotificationChannelIDs = nil
	m.HTTP.Headers = nil
	m.HTTP.Query = nil
	m.HTTP.Body.Fields = nil
	m.HTTP.Body.Files = nil
	m.HTTP.Assertions.Regex = nil
	ctx := context.Background()
	row, err := s.Store.GetMonitor(ctx, m.ID)
	if err != nil {
		t.Fatal(err)
	}
	row.ConfigJSON, err = json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Store.WithTx(ctx, func(tx *store.Tx) error { return tx.PutMonitor(ctx, row) }); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/api/v1/monitors/" + m.ID, "/api/v1/monitors"} {
		status, body := request(t, c, "GET", ts.URL+path, "", nil)
		if status != 200 {
			t.Fatalf("read monitor: %d %s", status, body)
		}
		var value map[string]any
		if err := json.Unmarshal(body, &value); err != nil {
			t.Fatal(err)
		}
		if path == "/api/v1/monitors" {
			value = value["items"].([]any)[0].(map[string]any)
		}
		for _, keys := range [][]string{
			{"tags"}, {"notificationChannelIds"}, {"http", "headers"},
			{"http", "query"}, {"http", "body", "fields"},
			{"http", "body", "files"}, {"http", "assertions", "regex"},
		} {
			var item any = value
			for _, key := range keys {
				item = item.(map[string]any)[key]
			}
			if array, ok := item.([]any); !ok || len(array) != 0 {
				t.Fatalf("%s %v must be an empty array, got %#v", path, keys, item)
			}
		}
	}
}

func TestAPIJSONPreservesNullableValues(t *testing.T) {
	s := New(nil, nil, config.Config{})
	value := struct {
		Groups       []domain.PageGroup  `json:"groups"`
		Tags         []string            `json:"tags"`
		Availability domain.Availability `json:"availability"`
		Raw          json.RawMessage     `json:"raw"`
		RawNull      json.RawMessage     `json:"rawNull"`
		RawNil       json.RawMessage     `json:"rawNil"`
		Metadata     map[string]string   `json:"metadata"`
		Published    *domain.PageConfig  `json:"published,omitempty"`
		Enabled      bool                `json:"enabled"`
		Label        string              `json:"label"`
	}{
		Groups:  []domain.PageGroup{{ID: "group", Name: "Services"}},
		Tags:    []string{"production"},
		Raw:     json.RawMessage(`{"values":null}`),
		RawNull: json.RawMessage(`null`),
		Label:   "<service>",
	}
	for _, contentType := range []string{"application/json", "application/problem+json"} {
		var body bytes.Buffer
		if err := s.API.Marshal(&body, contentType, value); err != nil {
			t.Fatal(err)
		}
		var got map[string]any
		if err := json.Unmarshal(body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		groups := got["groups"].([]any)
		if monitors := groups[0].(map[string]any)["monitors"]; !reflect.DeepEqual(monitors, []any{}) {
			t.Fatalf("nested collection is not an empty array: %s", body.Bytes())
		}
		if !reflect.DeepEqual(got["tags"], []any{"production"}) || got["enabled"] != false {
			t.Fatalf("existing values changed: %s", body.Bytes())
		}
		stats := got["availability"].(map[string]any)
		for _, entry := range []struct {
			object map[string]any
			key    string
		}{{stats, "uptime"}, {stats, "coverage"}, {got, "rawNull"}, {got, "rawNil"}, {got, "metadata"}} {
			if item, present := entry.object[entry.key]; !present || item != nil {
				t.Fatalf("meaningful null value %s changed: %s", entry.key, body.Bytes())
			}
		}
		if !reflect.DeepEqual(got["raw"], map[string]any{"values": nil}) {
			t.Fatalf("raw JSON changed: %s", body.Bytes())
		}
		if _, present := got["published"]; present || !bytes.Contains(body.Bytes(), []byte(`"label":"<service>"`)) {
			t.Fatalf("omitempty or HTML escaping changed: %s", body.Bytes())
		}
	}
}

func TestAPICollectionSchemas(t *testing.T) {
	s := New(nil, nil, config.Config{})
	seen := map[*huma.Schema]bool{}
	var check func(*huma.Schema)
	check = func(schema *huma.Schema) {
		if schema == nil || seen[schema] {
			return
		}
		seen[schema] = true
		if schema.Type == "array" && schema.Nullable {
			t.Errorf("collection schema allows null: %+v", schema)
		}
		check(schema.Items)
		for _, property := range schema.Properties {
			check(property)
		}
	}
	for _, schema := range s.API.OpenAPI().Components.Schemas.Map() {
		check(schema)
	}
	monitor := s.API.OpenAPI().Components.Schemas.Map()["Monitor"]
	for _, key := range []string{"tags", "notificationChannelIds"} {
		for _, required := range monitor.Required {
			if required == key {
				t.Errorf("optional monitor field %s became required", key)
			}
		}
	}
	availability := s.API.OpenAPI().Components.Schemas.Map()["Availability"]
	if !availability.Properties["uptime"].Nullable || !availability.Properties["coverage"].Nullable {
		t.Fatal("statistics must remain nullable")
	}
}
