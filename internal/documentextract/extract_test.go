package documentextract

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeZipFile(t *testing.T, path string, files map[string]string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	for name, body := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = w.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestExtractOfficeFormatsInOrder(t *testing.T) {
	dir := t.TempDir()
	docx := filepath.Join(dir, "report.docx")
	writeZipFile(t, docx, map[string]string{"word/document.xml": `<w:document xmlns:w="urn"><w:body><w:p><w:r><w:t>Intro</w:t></w:r></w:p><w:tbl><w:tr><w:tc><w:p><w:r><w:t>A</w:t></w:r></w:p></w:tc><w:tc><w:p><w:r><w:t>B</w:t></w:r></w:p></w:tc></w:tr></w:tbl></w:body></w:document>`})
	docText, err := Extract(docx, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(docText, "Intro") || !strings.Contains(docText, "A") || !strings.Contains(docText, "B") {
		t.Fatalf("docx text = %q", docText)
	}

	xlsx := filepath.Join(dir, "book.xlsx")
	writeZipFile(t, xlsx, map[string]string{
		"xl/workbook.xml":            `<workbook xmlns:r="urn"><sheets><sheet name="Sheet One" r:id="rId1"/></sheets></workbook>`,
		"xl/_rels/workbook.xml.rels": `<Relationships><Relationship Id="rId1" Target="worksheets/sheet1.xml"/></Relationships>`,
		"xl/sharedStrings.xml":       `<sst><si><t>Shared</t></si></sst>`,
		"xl/worksheets/sheet1.xml":   `<worksheet><sheetData><row><c r="A1" t="s"><v>0</v></c><c r="B1"><f>SUM(A1)</f><v>2</v></c></row></sheetData></worksheet>`,
	})
	xlsxText, err := Extract(xlsx, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(xlsxText, "Shared") || !strings.Contains(xlsxText, "SUM(A1)") {
		t.Fatalf("xlsx text = %q", xlsxText)
	}

	pptx := filepath.Join(dir, "slides.pptx")
	writeZipFile(t, pptx, map[string]string{
		"ppt/slides/slide2.xml": `<p:sld xmlns:a="urn"><a:t>Second</a:t></p:sld>`,
		"ppt/slides/slide1.xml": `<p:sld xmlns:a="urn"><a:t>First</a:t></p:sld>`,
	})
	pptText, err := Extract(pptx, 0)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Index(pptText, "First") > strings.Index(pptText, "Second") {
		t.Fatalf("pptx order = %q", pptText)
	}
}

func TestUnsupportedOfficeFormatIsExplicit(t *testing.T) {
	if IsSupportedOffice("legacy.doc") || IsSupportedOffice("legacy.xls") {
		t.Fatal("legacy binary Office format should remain unsupported")
	}
}
