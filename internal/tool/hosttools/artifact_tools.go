package hosttools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/artifact"
	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/tool"
)

// ArtifactTools returns the bundled Work artifact surface. It is registered by
// profile-aware boot code for Assistant and Orca only.
func ArtifactTools(workDir string) []tool.Tool {
	return []tool.Tool{
		artifactCreateTool{workDir: workDir}, artifactEditTool{workDir: workDir},
		artifactPreviewTool{workDir: workDir}, artifactValidateTool{workDir: workDir},
	}
}

func artifactPath(root, path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("path is required")
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	target := path
	if !filepath.IsAbs(target) {
		target = filepath.Join(rootAbs, target)
	}
	target, err = filepath.Abs(target)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(rootAbs, target)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("artifact path is outside the workspace")
	}
	return target, nil
}

type artifactCreateTool struct{ workDir string }

func (artifactCreateTool) Name() string { return "artifact_create" }
func (artifactCreateTool) Description() string {
	return "Create a structured DOCX, XLSX, PPTX, or PDF with a sidecar for reliable later editing; validates the result before returning."
}
func (artifactCreateTool) ReadOnly() bool { return false }
func (artifactCreateTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"},"format":{"type":"string","enum":["docx","xlsx","pptx","pdf"]},"title":{"type":"string"},"paragraphs":{"type":"array","items":{"type":"string"}},"sheets":{"type":"array","items":{"type":"object","properties":{"name":{"type":"string"},"rows":{"type":"array","items":{"type":"array","items":{"type":"string"}}}}}},"slides":{"type":"array","items":{"type":"object","properties":{"title":{"type":"string"},"bullets":{"type":"array","items":{"type":"string"}}}}}},"required":["path","format"]}`)
}
func (t artifactCreateTool) Execute(_ context.Context, raw json.RawMessage) (string, error) {
	var p struct {
		Path, Format, Title string
		Paragraphs          []string
		Sheets              []artifact.WorkbookSheet
		Slides              []artifact.Slide
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return "", err
	}
	path, err := artifactPath(t.workDir, p.Path)
	if err != nil {
		return "", err
	}
	result, err := artifact.Create(path, artifact.Model{Format: p.Format, Title: p.Title, Paragraphs: p.Paragraphs, Sheets: p.Sheets, Slides: p.Slides})
	if err != nil {
		return "", err
	}
	b, _ := json.Marshal(map[string]any{"status": "done", "path": path, "sidecar": artifact.SidecarPath(path), "validation": result})
	return string(b), nil
}

type artifactEditTool struct{ workDir string }

func (artifactEditTool) Name() string { return "artifact_edit" }
func (artifactEditTool) Description() string {
	return "Edit an Orca-created artifact structurally and regenerate and validate the file."
}
func (artifactEditTool) ReadOnly() bool { return false }
func (artifactEditTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"},"operations":{"type":"array","items":{"type":"object","properties":{"op":{"type":"string","enum":["append_text","replace_text","set_cell","add_slide"]},"text":{"type":"string"},"find":{"type":"string"},"replace":{"type":"string"},"sheet":{"type":"integer","minimum":0},"row":{"type":"integer","minimum":0},"column":{"type":"integer","minimum":0},"title":{"type":"string"},"bullets":{"type":"array","items":{"type":"string"}}},"required":["op"]}}},"required":["path","operations"]}`)
}
func (t artifactEditTool) Execute(_ context.Context, raw json.RawMessage) (string, error) {
	type operation struct {
		Op, Text, Find, Replace, Title string
		Sheet, Row, Column             int
		Bullets                        []string
	}
	var p struct {
		Path       string
		Operations []operation
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return "", err
	}
	path, err := artifactPath(t.workDir, p.Path)
	if err != nil {
		return "", err
	}
	result, err := artifact.Edit(path, func(m *artifact.Model) error {
		for _, op := range p.Operations {
			switch op.Op {
			case "append_text":
				m.Paragraphs = append(m.Paragraphs, op.Text)
			case "replace_text":
				for i, v := range m.Paragraphs {
					m.Paragraphs[i] = strings.ReplaceAll(v, op.Find, op.Replace)
				}
				for si := range m.Sheets {
					for ri := range m.Sheets[si].Rows {
						for ci := range m.Sheets[si].Rows[ri] {
							m.Sheets[si].Rows[ri][ci] = strings.ReplaceAll(m.Sheets[si].Rows[ri][ci], op.Find, op.Replace)
						}
					}
				}
				for si := range m.Slides {
					m.Slides[si].Title = strings.ReplaceAll(m.Slides[si].Title, op.Find, op.Replace)
					for bi := range m.Slides[si].Bullets {
						m.Slides[si].Bullets[bi] = strings.ReplaceAll(m.Slides[si].Bullets[bi], op.Find, op.Replace)
					}
				}
			case "set_cell":
				if op.Sheet < 0 || op.Sheet >= len(m.Sheets) {
					return fmt.Errorf("sheet index out of range")
				}
				for len(m.Sheets[op.Sheet].Rows) <= op.Row {
					m.Sheets[op.Sheet].Rows = append(m.Sheets[op.Sheet].Rows, []string{})
				}
				for len(m.Sheets[op.Sheet].Rows[op.Row]) <= op.Column {
					m.Sheets[op.Sheet].Rows[op.Row] = append(m.Sheets[op.Sheet].Rows[op.Row], "")
				}
				m.Sheets[op.Sheet].Rows[op.Row][op.Column] = op.Text
			case "add_slide":
				m.Slides = append(m.Slides, artifact.Slide{Title: op.Title, Bullets: op.Bullets})
			default:
				return fmt.Errorf("unsupported edit operation %q", op.Op)
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	b, _ := json.Marshal(map[string]any{"status": "done", "path": path, "validation": result})
	return string(b), nil
}

type artifactPreviewTool struct{ workDir string }

func (artifactPreviewTool) Name() string { return "artifact_preview" }
func (artifactPreviewTool) Description() string {
	return "Generate a deterministic PNG layout preview for an Orca-created artifact."
}
func (artifactPreviewTool) ReadOnly() bool { return false }
func (artifactPreviewTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"},"output":{"type":"string"}},"required":["path"]}`)
}
func (t artifactPreviewTool) Execute(_ context.Context, raw json.RawMessage) (string, error) {
	var p struct{ Path, Output string }
	if err := json.Unmarshal(raw, &p); err != nil {
		return "", err
	}
	path, err := artifactPath(t.workDir, p.Path)
	if err != nil {
		return "", err
	}
	output := ""
	if strings.TrimSpace(p.Output) != "" {
		output, err = artifactPath(t.workDir, p.Output)
		if err != nil {
			return "", err
		}
	}
	result, err := artifact.Preview(path, output)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("status=done preview=%s", result), nil
}

type artifactValidateTool struct{ workDir string }

func (artifactValidateTool) Name() string { return "artifact_validate" }
func (artifactValidateTool) Description() string {
	return "Re-open and validate an Orca-created DOCX, XLSX, PPTX, or PDF and report its logical unit and text counts."
}
func (artifactValidateTool) ReadOnly() bool { return true }
func (artifactValidateTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"}},"required":["path"]}`)
}
func (t artifactValidateTool) Execute(_ context.Context, raw json.RawMessage) (string, error) {
	var p struct{ Path string }
	if err := json.Unmarshal(raw, &p); err != nil {
		return "", err
	}
	path, err := artifactPath(t.workDir, p.Path)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(path); err != nil {
		return "", err
	}
	result, err := artifact.Validate(path)
	if err != nil {
		return "", err
	}
	b, _ := json.Marshal(result)
	return string(b), nil
}
