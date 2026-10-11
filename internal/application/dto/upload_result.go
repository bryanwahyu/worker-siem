package dto

// UploadResult communicates the outcome of a batch upload.
type UploadResult struct {
	Uploaded int
	Failed   int
	Spooled  int
}
