package service

import "encoding/json"

// TemplateService provides document templates.
type TemplateService interface {
	GetTemplates() []*DocumentTemplate
	GetTemplate(id string) *DocumentTemplate
	GetTemplateBlocks(id string) []TemplateBlock
}

// DocumentTemplate represents a document template.
type DocumentTemplate struct {
	ID     string          `json:"id"`
	Name   string          `json:"name"`
	Icon   string          `json:"icon"`
	Blocks []TemplateBlock `json:"blocks"`
}

// TemplateBlock represents a block within a template.
type TemplateBlock struct {
	Type    string          `json:"type"`
	Content string          `json:"content"`
	Rows    json.RawMessage `json:"rows,omitempty"`
	Columns json.RawMessage `json:"columns,omitempty"`
	Emoji   string          `json:"emoji,omitempty"`
	Color   string          `json:"color,omitempty"`
}

type templateService struct {
	templates map[string]*DocumentTemplate
	ordered   []*DocumentTemplate
}

// NewTemplateService creates a new template service with predefined templates.
func NewTemplateService() TemplateService {
	templates := buildTemplates()
	m := make(map[string]*DocumentTemplate, len(templates))
	for _, t := range templates {
		m[t.ID] = t
	}
	return &templateService{
		templates: m,
		ordered:   templates,
	}
}

func (s *templateService) GetTemplates() []*DocumentTemplate {
	return s.ordered
}

func (s *templateService) GetTemplate(id string) *DocumentTemplate {
	return s.templates[id]
}

func (s *templateService) GetTemplateBlocks(id string) []TemplateBlock {
	t := s.templates[id]
	if t == nil {
		return nil
	}
	return t.Blocks
}

func buildTemplates() []*DocumentTemplate {
	return []*DocumentTemplate{
		{
			ID:   "kosong",
			Name: "Dokumen Kosong",
			Icon: "\U0001F4C4", // 📄
			Blocks: []TemplateBlock{
				{Type: "paragraph", Content: ""},
			},
		},
		{
			ID:   "notulen-rapat",
			Name: "Notulen Rapat",
			Icon: "\U0001F4CB", // 📋
			Blocks: []TemplateBlock{
				{Type: "heading1", Content: "Notulen Rapat"},
				{Type: "heading2", Content: "Agenda"},
				{Type: "numbered-list", Content: ""},
				{Type: "heading2", Content: "Peserta"},
				{Type: "bullet-list", Content: ""},
				{Type: "heading2", Content: "Pembahasan"},
				{Type: "paragraph", Content: ""},
				{Type: "heading2", Content: "Keputusan"},
				{Type: "numbered-list", Content: ""},
				{Type: "heading2", Content: "Tindak Lanjut"},
				{Type: "checklist", Content: ""},
			},
		},
		{
			ID:   "daftar-belanja",
			Name: "Daftar Belanja",
			Icon: "\U0001F6D2", // 🛒
			Blocks: []TemplateBlock{
				{Type: "heading1", Content: "Daftar Belanja"},
				{
					Type: "table",
					Columns: mustJSON([]map[string]string{
						{"id": "col-1", "name": "Nama Barang"},
						{"id": "col-2", "name": "Jumlah"},
						{"id": "col-3", "name": "Harga Satuan"},
						{"id": "col-4", "name": "Total"},
					}),
					Rows: mustJSON([]map[string]string{
						{"col-1": "", "col-2": "", "col-3": "", "col-4": ""},
					}),
				},
			},
		},
		{
			ID:   "catatan-keuangan",
			Name: "Catatan Keuangan",
			Icon: "\U0001F4B0", // 💰
			Blocks: []TemplateBlock{
				{Type: "heading1", Content: "Catatan Keuangan"},
				{
					Type: "table",
					Columns: mustJSON([]map[string]string{
						{"id": "col-1", "name": "Tanggal"},
						{"id": "col-2", "name": "Keterangan"},
						{"id": "col-3", "name": "Pemasukan"},
						{"id": "col-4", "name": "Pengeluaran"},
						{"id": "col-5", "name": "Saldo"},
					}),
					Rows: mustJSON([]map[string]string{
						{"col-1": "", "col-2": "", "col-3": "", "col-4": "", "col-5": ""},
					}),
				},
			},
		},
		{
			ID:   "catatan-kesehatan",
			Name: "Catatan Kesehatan",
			Icon: "\U0001FA7A", // 🩺
			Blocks: []TemplateBlock{
				{Type: "heading1", Content: "Catatan Kesehatan"},
				{Type: "heading2", Content: "Keluhan"},
				{Type: "paragraph", Content: ""},
				{Type: "heading2", Content: "Diagnosis"},
				{Type: "paragraph", Content: ""},
				{Type: "heading2", Content: "Obat"},
				{Type: "bullet-list", Content: ""},
				{Type: "heading2", Content: "Dokter"},
				{Type: "paragraph", Content: ""},
				{Type: "heading2", Content: "Kunjungan Berikutnya"},
				{Type: "paragraph", Content: ""},
			},
		},
		{
			ID:   "kesepakatan-bersama",
			Name: "Kesepakatan Bersama",
			Icon: "\U0001F91D", // 🤝
			Blocks: []TemplateBlock{
				{Type: "heading1", Content: "Kesepakatan Bersama"},
				{Type: "heading2", Content: "Pihak-Pihak"},
				{Type: "numbered-list", Content: ""},
				{Type: "heading2", Content: "Isi Kesepakatan"},
				{Type: "paragraph", Content: ""},
				{Type: "heading2", Content: "Ketentuan"},
				{Type: "numbered-list", Content: ""},
				{Type: "divider"},
				{Type: "heading2", Content: "Tanda Tangan"},
				{Type: "paragraph", Content: ""},
			},
		},
		{
			ID:   "catatan-pertanian",
			Name: "Catatan Pertanian",
			Icon: "\U0001F33E", // 🌾
			Blocks: []TemplateBlock{
				{Type: "heading1", Content: "Catatan Pertanian"},
				{
					Type: "table",
					Columns: mustJSON([]map[string]string{
						{"id": "col-1", "name": "Lahan"},
						{"id": "col-2", "name": "Tanaman"},
						{"id": "col-3", "name": "Tanggal Tanam"},
						{"id": "col-4", "name": "Hasil Panen"},
						{"id": "col-5", "name": "Catatan"},
					}),
					Rows: mustJSON([]map[string]string{
						{"col-1": "", "col-2": "", "col-3": "", "col-4": "", "col-5": ""},
					}),
				},
			},
		},
		{
			ID:   "inventaris-aset",
			Name: "Inventaris Aset",
			Icon: "\U0001F4E6", // 📦
			Blocks: []TemplateBlock{
				{Type: "heading1", Content: "Inventaris Aset"},
				{
					Type: "table",
					Columns: mustJSON([]map[string]string{
						{"id": "col-1", "name": "Nama Aset"},
						{"id": "col-2", "name": "Jenis"},
						{"id": "col-3", "name": "Lokasi"},
						{"id": "col-4", "name": "Kondisi"},
						{"id": "col-5", "name": "Catatan"},
					}),
					Rows: mustJSON([]map[string]string{
						{"col-1": "", "col-2": "", "col-3": "", "col-4": "", "col-5": ""},
					}),
				},
			},
		},
	}
}

func mustJSON(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
