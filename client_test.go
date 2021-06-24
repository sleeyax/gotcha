package gotcha

import "testing"

func TestClient_Extend(t *testing.T) {
	client, err := NewClient(&Options{
		PrefixURL: "https://before.example.com",
	})
	if err != nil {
		t.Fatal(err)
	}

	prefixUrl := "https://after.example.com"

	client2, err := client.Extend(&Options{
		PrefixURL: prefixUrl,
	})
	if err != nil {
		t.Fatal(err)
	}

	if client2.options.PrefixURL != prefixUrl {
		t.FailNow()
	}
}
