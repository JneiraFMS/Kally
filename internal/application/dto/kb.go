package dto

import domainkb "grz/internal/domain/kb"

// KBInput is the application input used to create or update a knowledge base.
type KBInput struct {
	ID         string
	Erreur     string
	Notes      string
	Resolution string
	Tags       string
}

// KBDetail is the detail page data prepared by the application layer.
type KBDetail struct {
	KB      domainkb.KB
	Related []domainkb.KB
}
