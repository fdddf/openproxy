package services

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestParseChatGPTAccountID(t *testing.T) {
	buildJWT := func(payload map[string]interface{}) string {
		header := map[string]interface{}{"alg": "none", "typ": "JWT"}
		headerBytes, _ := json.Marshal(header)
		payloadBytes, _ := json.Marshal(payload)

		return base64.RawURLEncoding.EncodeToString(headerBytes) + "." +
			base64.RawURLEncoding.EncodeToString(payloadBytes) + ".signature"
	}
	svc := &codexOAuthService{}

	tests := []struct {
		name    string
		token   string
		want    string
		wantErr bool
	}{
		{
			name: "account id present",
			token: buildJWT(map[string]interface{}{
				"https://api.openai.com/auth": map[string]interface{}{
					"chatgpt_account_id": "x-8e72-47eb-x-c",
				},
			}),
			want: "x-8e72-47eb-x-c",
		},
		{
			name: "claim missing returns empty",
			token: buildJWT(map[string]interface{}{
				"https://api.openai.com/profile": map[string]interface{}{
					"email": "test@example.com",
				},
			}),
			want: "",
		},
		{
			name:    "invalid jwt format",
			token:   "not-a-jwt",
			wantErr: true,
		},
		{
			name:    "invalid base64 payload",
			token:   "header.invalid_payload@@.signature",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.parseChatGPTAccountID(tt.token)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseChatGPTAccountID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("parseChatGPTAccountID() got = %q, want %q", got, tt.want)
			}
		})
	}
}