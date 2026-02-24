package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	wtConfigDir = "configs/microservice/bff-service/configs/workflow-template"
)

type WorkflowTemplateConfig struct {
	TemplateId string `json:"templateId" mapstructure:"templateId"`
	Category   string `json:"category" mapstructure:"category"`
	Avatar     string `json:"avatar"`
	Name       string `json:"name"`
	SchemaPath string `json:"schemaPath" mapstructure:"schemaPath"`
	Schema     []byte `json:"-" mapstructure:"-"`
	Desc       string `json:"desc" mapstructure:"desc"`
	Author     string `json:"author" mapstructure:"author"`
	Summary    string `json:"summary" mapstructure:"summary"`
	Feature    string `json:"feature" mapstructure:"feature"`
	Scenario   string `json:"scenario" mapstructure:"scenario"`
	Note       string `json:"note" mapstructure:"note"`
}

func (wtf *WorkflowTemplateConfig) load() error {
	schemaPath := filepath.Join(wtConfigDir, wtf.SchemaPath)
	b, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("load workflowtemp %v schema path %v err: %v", wtf.TemplateId, schemaPath, err)
	}
	wtf.Schema = b
	// avatarPath := filepath.Join(ConfigDir, wtf.Avatar)
	// if _, err = os.ReadFile(avatarPath); err != nil {
	//     return fmt.Errorf("load workflow %v avatar path %v err: %v", wtf.TemplateId, avatarPath, err)
	// }
	return nil
}
