package compress

// File size constants
const (
	// Sizes in bytes
	MB int64 = 1024 * 1024
	GB int64 = 1024 * MB

	// Standard tar format limit is approximately 8GB
	// This is the limit for a single file in standard tar format
	StandardTarSizeLimit int64 = 8 * GB

	// Recommended max file size (slightly less than the actual limit for safety)
	// Using PAX format will allow files larger than this limit
	RecommendedMaxFileSize int64 = StandardTarSizeLimit - 100*MB
)
