package registry

// DocumentV2 is the runtime registry protocol published by kuaimai-cli-open.
type DocumentV2 struct {
	SchemaVersion string                       `json:"schemaVersion"`
	Version       string                       `json:"version"`
	GeneratedAt   string                       `json:"generatedAt"`
	Source        string                       `json:"source"`
	Domains       map[string]DomainIndex       `json:"domains"`
	APIs          map[string]APIEntry          `json:"apis"`
}

// DomainIndex groups apis by business domain.
type DomainIndex struct {
	Label  string   `json:"label"`
	Count  int      `json:"count"`
	APIIDs []string `json:"apiIds"`
}

// APIEntry describes one callable web API.
type APIEntry struct {
	ID              string         `json:"id"`
	Domain          string         `json:"domain"`
	Title           string         `json:"title"`
	Description     string         `json:"description"`
	Transport       string         `json:"transport"`
	Method          string         `json:"method"`
	Path            string         `json:"path"`
	BaseURL         string         `json:"baseUrl"`
	ContentType     string         `json:"contentType"`
	Risk            string         `json:"risk"`
	Write           bool           `json:"write"`
	Pageable        bool           `json:"pageable"`
	Stability       string         `json:"stability"`
	Auth            map[string]any `json:"auth"`
	RequestSchema   *JSONSchema    `json:"requestSchema"`
	ResponseSchema  *JSONSchema    `json:"responseSchema"`
	Response        map[string]any `json:"response"`
	Source          map[string]any `json:"source"`
	Examples        []map[string]any `json:"examples"`
}
