package sdk

// CreateSnapshotRequest is representation of a snapshot request.
type CreateSnapshotRequest struct {
	Expires   uint  `json:"expires"`
	Dashboard Board `json:"dashboard"`
}
