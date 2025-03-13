package types

var (
	K8sAgentServiceURL string // Flag to store the k8s-agent-service LoadBalancer IP
	AiAgent            string // Flag to use the Ai Agent { Gemini, Cohere, Deepseek etc. }
	AiAgentKey         string // Flag to store the Ai Agent ApiKey
)

// Alert struct with the expected parameters
type Alert struct {
	Description     string `json:"description"`
	RemediationYAML string `json:"remediationYAML"`
}
