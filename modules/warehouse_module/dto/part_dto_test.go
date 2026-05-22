package dto

import (
	"testing"

	"github.com/gin-gonic/gin/binding"
)

func TestPartCreateRequestBindingAllowsZeroStartQuantity(t *testing.T) {
	body := []byte(`{"part_id":"BRK-PAD-001","name":"Brake Pad Front","start_quantity":0,"category":"brake_system"}`)

	var req PartCreateRequest
	if err := binding.JSON.BindBody(body, &req); err != nil {
		t.Fatalf("expected start_quantity=0 to bind successfully, got error: %v", err)
	}
}

func TestPartCreateRequestBindingRejectsNegativeStartQuantity(t *testing.T) {
	body := []byte(`{"part_id":"BRK-PAD-001","name":"Brake Pad Front","start_quantity":-1,"category":"brake_system"}`)

	var req PartCreateRequest
	if err := binding.JSON.BindBody(body, &req); err == nil {
		t.Fatal("expected negative start_quantity to fail validation")
	}
}
