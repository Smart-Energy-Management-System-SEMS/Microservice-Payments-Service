package acl

import "testing"

func TestTranslateSubscriptionCreatedSupportsCurrentEnvelope(t *testing.T) {
	payload := []byte(`{"eventType":"subscription.created","data":{"subscription_id":"sub-1","user_id":"user-1"}}`)

	event, err := TranslateSubscriptionCreated(payload)
	if err != nil {
		t.Fatalf("TranslateSubscriptionCreated returned error: %v", err)
	}
	if event.SubscriptionID != "sub-1" {
		t.Fatalf("SubscriptionID = %q, want %q", event.SubscriptionID, "sub-1")
	}
	if event.UserID != "user-1" {
		t.Fatalf("UserID = %q, want %q", event.UserID, "user-1")
	}
}

func TestTranslateSubscriptionCreatedSupportsLegacyFieldsInsideData(t *testing.T) {
	payload := []byte(`{"eventType":"subscription.created","data":{"SubscriptionID":"sub-legacy","UserID":"user-legacy"}}`)

	event, err := TranslateSubscriptionCreated(payload)
	if err != nil {
		t.Fatalf("TranslateSubscriptionCreated returned error: %v", err)
	}
	if event.SubscriptionID != "sub-legacy" {
		t.Fatalf("SubscriptionID = %q, want %q", event.SubscriptionID, "sub-legacy")
	}
	if event.UserID != "user-legacy" {
		t.Fatalf("UserID = %q, want %q", event.UserID, "user-legacy")
	}
}

func TestTranslateSubscriptionCreatedSupportsTopLevelLegacyPayload(t *testing.T) {
	payload := []byte(`{"eventType":"subscription.created","SubscriptionID":"sub-top","UserID":"user-top"}`)

	event, err := TranslateSubscriptionCreated(payload)
	if err != nil {
		t.Fatalf("TranslateSubscriptionCreated returned error: %v", err)
	}
	if event.SubscriptionID != "sub-top" {
		t.Fatalf("SubscriptionID = %q, want %q", event.SubscriptionID, "sub-top")
	}
	if event.UserID != "user-top" {
		t.Fatalf("UserID = %q, want %q", event.UserID, "user-top")
	}
}
