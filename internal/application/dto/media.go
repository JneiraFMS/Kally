package dto

// Media is a stored image reference returned after a successful upload. URL
// is the path the presentation layer inserts into the markdown.
type Media struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// MediaFile is a stored image read back for serving, with the MIME type
// derived from its content-addressed name.
type MediaFile struct {
	Name    string
	Content []byte
	MIME    string
}
