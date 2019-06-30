package models

type JsonData map[string]interface{}

type Entry struct {
	ID    int      `db:"entry_id"`
	Acc   string   `db:"acc"`
	TaxID int      `db:"tax_id"`
	Data  JsonData `db:"data"`
}

type StatCluster struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Count       int    `json:"count"`
	Class       string `json:"css_class"`
}

type ProductTag struct {
	Name  string `json:"name"`
	Class string `json:"css_class"`
}

type Compound struct {
	Name string `json:"compound"`
}

type CompoundList []Compound

type RepositoryEntry struct {
	Accession    string       `json:"accession"`
	Minimal      bool         `json:"minimal"`
	Products     []string     `json:"products"`
	ProductTags  []ProductTag `json:"classes"`
	OrganismName string       `json:"organism"`
}

type AccessionRequestLocus struct {
	GenBankAccession string `json:"genbank_accession"`
	Start            int    `json:"start"`
	End              int    `json:"end"`
}

type AccessionRequest struct {
	Name      string                  `json:"name"`
	Email     string                  `json:"email"`
	Compounds []string                `json:"compounds"`
	Loci      []AccessionRequestLocus `json:"loci"`
}
