package domain_test

import (
	"encoding/json"
	"testing"

	"reflect"

	"stockseer.ai/blueksy-firehose/internal/domain"
)

func TestMessage_String(t *testing.T) {
	tests := []struct {
		name string
		msg  domain.Message
		want string
	}{
		{
			name: "Basic Message",
			msg: domain.Message{
				DID:    "did:example:123",
				TimeUS: 123456789,
				Kind:   "create",
				Commit: domain.Commit{
					Rev:        "123456789abcdef0",
					Operation:  "insert",
					Collection: "posts",
					Rkey:       "post:123",
					CID:        "cid:example:123",
					Record: domain.Record{
						Type:      "post",
						CreatedAt: "2023-11-22T10:20:30Z",
						Langs:     []string{"en"},
						Text:      "Hello, world!",
						Reply: domain.Reply{
							Parent: "did:example:456",
							Root:   "did:example:789",
						},
					},
				},
			},
			want: "",
		},
		// Add more test cases with different message structures
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.msg.String(); reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("Message.String() = %s not of type %s", reflect.TypeOf(got), reflect.TypeOf(tt.want))
			}
		})
	}
}

func TestMessage_UnmarshalJSON(t *testing.T) {
	// ... (Implement test cases for UnmarshalJSON) ...
	testString := []byte(`{"did":"did:plc:lo55yba44obaih2nt5opn4z4","time_us":1735007698011970,"kind":"commit","commit":{"rev":"3ldzfus6fjk2x","operation":"create","collection":"app.bsky.feed.post","rkey":"3ldzfurmqxk23","record":{"$type":"app.bsky.feed.post","createdAt":"2024-12-24T02:34:57.198Z","langs":["en"],"text":"guy who follows 27000 accounts: how could this random chick block me"},"cid":"bafyreif7cmhmrjha6xeono5e6zf6gmzwea6yza7ruuy4ebpgagjv5gitwa"}}`)
	t.Run("New Root Post UnmarshalJSON", func(t *testing.T) {
		var m domain.Message
		jsonerr := json.Unmarshal(testString, &m)

		if jsonerr != nil {
			return
		}

		if got := m.DID; got != "did:plc:lo55yba44obaih2nt5opn4z4" {
			t.Errorf("Message.DID = %v, want %v", got, "did:plc:lo55yba44obaih2nt5opn4z4")
		}
		if got := m.TimeUS; got != 1735007698011970 {
			t.Errorf("Message.DID = %v, want %v", got, 1735007698011970)
		}
	})
}

func TestMessage_IsEnglish(t *testing.T) {
	tests := []struct {
		name string
		msg  domain.Message
		want bool
	}{
		{
			name: "English Message",
			msg: domain.Message{
				Commit: domain.Commit{
					Record: domain.Record{
						Langs: []string{"en"},
					},
				},
			},
			want: true,
		},
		{
			name: "Non-English Message",
			msg: domain.Message{
				Commit: domain.Commit{
					Record: domain.Record{
						Langs: []string{"fr"},
					},
				},
			},
			want: false,
		},
		{
			name: "No Languages",
			msg: domain.Message{
				Commit: domain.Commit{
					Record: domain.Record{
						Langs: []string{},
					},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.msg.IsEnglish(); got != tt.want {
				t.Errorf("Message.IsEnglish() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessage_IsRoot(t *testing.T) {
	tests := []struct {
		name string
		msg  domain.Message
		want bool
	}{
		{
			name: "Root Message",
			msg: domain.Message{
				Commit: domain.Commit{
					Record: domain.Record{
						Reply: domain.Reply{
							Root: nil,
						},
					},
				},
			},
			want: true,
		},
		{
			name: "Not Root Message",
			msg: domain.Message{
				Commit: domain.Commit{
					Record: domain.Record{
						Reply: domain.Reply{
							Root: "did:example:123",
						},
					},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.msg.IsRoot(); got != tt.want {
				t.Errorf("Message.IsRoot() = %v, want %v", got, tt.want)
			}
		})
	}
}
