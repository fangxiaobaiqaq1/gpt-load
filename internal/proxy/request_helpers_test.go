package proxy

import (
	"encoding/json"
	"testing"

	"gpt-load/internal/models"
)

func TestApplyStripParams_RemovesConfiguredTopLevelFields(t *testing.T) {
	ps := &ProxyServer{}
	group := &models.Group{StripParamsList: []string{"top_p", "presence_penalty"}}
	body := []byte(`{"model":"gpt-4.1","top_p":0.9,"presence_penalty":0.5,"messages":[{"role":"user","content":"hi"}]}`)

	got, err := ps.applyStripParams(body, group)
	if err != nil {
		t.Fatalf("applyStripParams returned error: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(got, &payload); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}

	if _, ok := payload["top_p"]; ok {
		t.Fatalf("expected top_p to be removed")
	}
	if _, ok := payload["presence_penalty"]; ok {
		t.Fatalf("expected presence_penalty to be removed")
	}
	if payload["model"] != "gpt-4.1" {
		t.Fatalf("expected model to remain, got %v", payload["model"])
	}
}

func TestApplyStripParams_PassesThroughOnInvalidJSON(t *testing.T) {
	ps := &ProxyServer{}
	group := &models.Group{StripParamsList: []string{"top_p"}}
	body := []byte(`{"model":`)

	got, err := ps.applyStripParams(body, group)
	if err != nil {
		t.Fatalf("applyStripParams returned error: %v", err)
	}
	if string(got) != string(body) {
		t.Fatalf("expected invalid JSON body to pass through unchanged")
	}
}

func TestApplyParamOverridesThenStripParams_StripsFinalValue(t *testing.T) {
	ps := &ProxyServer{}
	group := &models.Group{
		ParamOverrides:  map[string]any{"top_p": 0.1, "temperature": 0.7},
		StripParamsList: []string{"top_p"},
	}
	body := []byte(`{"model":"gpt-4.1","top_p":0.9,"temperature":0.2}`)

	overridden, err := ps.applyParamOverrides(body, group)
	if err != nil {
		t.Fatalf("applyParamOverrides returned error: %v", err)
	}
	got, err := ps.applyStripParams(overridden, group)
	if err != nil {
		t.Fatalf("applyStripParams returned error: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(got, &payload); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if _, ok := payload["top_p"]; ok {
		t.Fatalf("expected top_p to be removed after override step")
	}
	if payload["temperature"] != 0.7 {
		t.Fatalf("expected temperature override to remain, got %v", payload["temperature"])
	}
}
