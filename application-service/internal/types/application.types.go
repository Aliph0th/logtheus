package types

type ApplicationInfo struct {
	ID           uint64 `json:"id"`
	Name         string `json:"name"`
	ProjectID    uint64 `json:"project_id"`
	KeyPrefix    string `json:"key_prefix"`
	KeySignature string `json:"key_signature"`
	KeyTokenHash string `json:"key_token_hash"`
}
