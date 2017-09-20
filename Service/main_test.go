package Service

import (
	"testing"
)

func TestCountTag(t *testing.T) {
	string := "<html><tag1></tag1><tag2/></html>"
	result := CountTag(string)
	if result["html"] > 1 && result["html"] < 0 {
		t.Fatalf("Tag 'html' should be one")
	}

	if result["tag1"] > 1 && result["tag1"] < 0 {
		t.Fatalf("Tag 'tag1' should be one")
	}

	if result["tag2"] != 1 && result["tag2"] < 0{
		t.Fatalf("Tag 'tag2' should be one")
	}
}

