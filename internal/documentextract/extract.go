// Package documentextract extracts readable text from modern Office containers
// without requiring an external Office installation or Python runtime.
package documentextract

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const DefaultMaxBytes int64 = 1 << 20

func IsSupportedOffice(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".docx", ".xlsx", ".pptx":
		return true
	}
	return false
}

func Extract(path string, maxBytes int64) (string, error) {
	if maxBytes <= 0 {
		maxBytes = DefaultMaxBytes
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".docx":
		return extractDOCX(path, maxBytes)
	case ".xlsx":
		return extractXLSX(path, maxBytes)
	case ".pptx":
		return extractPPTX(path, maxBytes)
	case ".txt", ".md", ".json", ".csv", ".tsv", ".xml", ".html", ".log", ".go", ".ts", ".tsx", ".js", ".py":
		data, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		return truncate(string(data), maxBytes), nil
	default:
		return "", fmt.Errorf("unsupported document extension %q", filepath.Ext(path))
	}
}

func truncate(text string, max int64) string {
	if int64(len(text)) <= max {
		return text
	}
	return text[:max] + "\n[truncated]"
}

func zipEntry(path, entry string) ([]byte, error) {
	z, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer z.Close()
	for _, file := range z.File {
		if file.Name != entry {
			continue
		}
		r, err := file.Open()
		if err != nil {
			return nil, err
		}
		defer r.Close()
		return io.ReadAll(r)
	}
	return nil, fmt.Errorf("missing %s", entry)
}

func elementText(data []byte, accepted map[string]bool) string {
	decoder := xml.NewDecoder(strings.NewReader(string(data)))
	var stack []string
	var out strings.Builder
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return strings.TrimSpace(out.String())
		}
		switch value := token.(type) {
		case xml.StartElement:
			if value.Name.Local == "p" || value.Name.Local == "tr" {
				if out.Len() > 0 {
					out.WriteByte('\n')
				}
			}
			if value.Name.Local == "tc" && out.Len() > 0 {
				out.WriteByte('\t')
			}
			stack = append(stack, value.Name.Local)
		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case xml.CharData:
			if len(stack) > 0 && accepted[stack[len(stack)-1]] {
				out.Write(value)
			}
		}
	}
	return strings.TrimSpace(strings.ReplaceAll(out.String(), "\r", ""))
}

func extractDOCX(path string, max int64) (string, error) {
	data, err := zipEntry(path, "word/document.xml")
	if err != nil {
		return "", err
	}
	return truncate(elementText(data, map[string]bool{"t": true}), max), nil
}

func sharedStrings(data []byte) []string {
	decoder := xml.NewDecoder(strings.NewReader(string(data)))
	var values []string
	var inItem, inText bool
	var current strings.Builder
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		switch value := token.(type) {
		case xml.StartElement:
			if value.Name.Local == "si" {
				inItem = true
				current.Reset()
			}
			if value.Name.Local == "t" && inItem {
				inText = true
			}
		case xml.CharData:
			if inText {
				current.Write(value)
			}
		case xml.EndElement:
			if value.Name.Local == "t" {
				inText = false
			}
			if value.Name.Local == "si" && inItem {
				values = append(values, current.String())
				inItem = false
			}
		}
	}
	return values
}

func workbookSheets(data, rels []byte) []sheetRef {
	targets := map[string]string{}
	decoder := xml.NewDecoder(strings.NewReader(string(rels)))
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "Relationship" {
			continue
		}
		var id, target string
		for _, attr := range start.Attr {
			if attr.Name.Local == "Id" {
				id = attr.Value
			}
			if attr.Name.Local == "Target" {
				target = attr.Value
			}
		}
		if id != "" {
			targets[id] = "xl/" + strings.TrimPrefix(target, "/xl/")
		}
	}
	decoder = xml.NewDecoder(strings.NewReader(string(data)))
	var sheets []sheetRef
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "sheet" {
			continue
		}
		var name, id string
		for _, attr := range start.Attr {
			if attr.Name.Local == "name" {
				name = attr.Value
			}
			if attr.Name.Local == "id" {
				id = attr.Value
			}
		}
		sheets = append(sheets, sheetRef{name: name, path: targets[id]})
	}
	return sheets
}

type sheetRef struct{ name, path string }

func worksheetText(data []byte, shared []string) string {
	decoder := xml.NewDecoder(strings.NewReader(string(data)))
	var out strings.Builder
	var ref, kind, value, formula string
	inCell, inValue, inFormula := false, false, false
	flush := func() {
		if !inCell {
			return
		}
		display := value
		if kind == "s" {
			if index, err := strconv.Atoi(strings.TrimSpace(value)); err == nil && index >= 0 && index < len(shared) {
				display = shared[index]
			}
		}
		if formula != "" {
			display += " (=" + formula + ")"
		}
		if strings.TrimSpace(display) != "" {
			if out.Len() > 0 {
				out.WriteByte('\n')
			}
			out.WriteString(ref + ": " + display)
		}
	}
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		switch item := token.(type) {
		case xml.StartElement:
			switch item.Name.Local {
			case "c":
				inCell = true
				ref, kind, value, formula = "", "", "", ""
				for _, attr := range item.Attr {
					if attr.Name.Local == "r" {
						ref = attr.Value
					}
					if attr.Name.Local == "t" {
						kind = attr.Value
					}
				}
			case "v":
				inValue = true
			case "f":
				inFormula = true
			}
		case xml.CharData:
			if inCell && inValue {
				value += string(item)
			}
			if inCell && inFormula {
				formula += string(item)
			}
		case xml.EndElement:
			switch item.Name.Local {
			case "v":
				inValue = false
			case "f":
				inFormula = false
			case "c":
				flush()
				inCell = false
			}
		}
	}
	return out.String()
}

func extractXLSX(path string, max int64) (string, error) {
	workbook, err := zipEntry(path, "xl/workbook.xml")
	if err != nil {
		return "", err
	}
	rels, _ := zipEntry(path, "xl/_rels/workbook.xml.rels")
	sharedData, _ := zipEntry(path, "xl/sharedStrings.xml")
	var out strings.Builder
	for index, sheet := range workbookSheets(workbook, rels) {
		if sheet.path == "" {
			sheet.path = fmt.Sprintf("xl/worksheets/sheet%d.xml", index+1)
		}
		data, err := zipEntry(path, sheet.path)
		if err != nil {
			continue
		}
		if out.Len() > 0 {
			out.WriteByte('\n')
		}
		out.WriteString("## " + sheet.name + "\n" + worksheetText(data, sharedStrings(sharedData)))
	}
	return truncate(strings.TrimSpace(out.String()), max), nil
}

func extractPPTX(path string, max int64) (string, error) {
	z, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer z.Close()
	var files []*zip.File
	for _, file := range z.File {
		if strings.HasPrefix(file.Name, "ppt/slides/slide") && strings.HasSuffix(file.Name, ".xml") {
			files = append(files, file)
		}
	}
	sort.Slice(files, func(i, j int) bool { return slideNumber(files[i].Name) < slideNumber(files[j].Name) })
	var out strings.Builder
	for _, file := range files {
		r, err := file.Open()
		if err != nil {
			continue
		}
		data, _ := io.ReadAll(r)
		r.Close()
		if out.Len() > 0 {
			out.WriteByte('\n')
		}
		out.WriteString("## " + file.Name + "\n" + elementText(data, map[string]bool{"t": true}))
	}
	return truncate(strings.TrimSpace(out.String()), max), nil
}

func slideNumber(name string) int {
	n, _ := strconv.Atoi(strings.TrimPrefix(strings.TrimSuffix(filepath.Base(name), ".xml"), "slide"))
	return n
}
