package assets

import (
	"bytes"
	"embed"
	"text/template"
)

//go:embed prompts/* messages/*
var Files embed.FS

// GetPrompt returns the system prompt template
func GetPrompt() (string, error) {
	b, err := Files.ReadFile("prompts/system_prompt.txt")
	return string(b), err
}

// GetOllamaWarning returns the Ollama warning message
func GetOllamaWarning() (string, error) {
	b, err := Files.ReadFile("messages/ollama_warning.txt")
	return string(b), err
}

// GetInitSuccess returns the initialization success message
func GetInitSuccess() (string, error) {
	b, err := Files.ReadFile("messages/init_success.txt")
	return string(b), err
}

// RenderOllamaWarning renders the Ollama warning message with the provided context
func RenderOllamaWarning(url, model string) (string, error) {
	warningTmpl, err := GetOllamaWarning()
	if err != nil {
		return "", err
	}
	tmpl, err := template.New("warning").Parse(warningTmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, struct{ URL, Model string }{URL: url, Model: model}); err != nil {
		return "", err
	}
	return buf.String(), nil
}
