package utils

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
)

// CreateArtifactName create a determinist model artifact name from the version name
func CreateArtifactName(versionName string) string {
	return fmt.Sprintf("%s/binary-model", versionName)
}

func ParseTemplate(templateFS embed.FS) (*template.Template, error) {
	template, err := template.ParseFS(templateFS, "templates/*.yaml.tmpl")
	if err != nil {
		return nil, err
	}
	return template, err
}

// Marshal converts a struct object into a prettifyied json string
func Marshal(obj interface{}) string {
	b, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		log.Fatalf("error marshalling object: %v", obj)
	}
	return string(b)
}

// Of returns a pointer to the provided literal/const input
func Of[E any](e E) *E {
	return &e
}
