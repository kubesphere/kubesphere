package sdk

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

// https://grafana.com/docs/grafana/latest/http_api/snapshot/

// CreateAnnotation creates a new snapshot.
func (r *Client) CreateSnapshot(ctx context.Context, a CreateSnapshotRequest) (StatusMessage, error) {
	var (
		raw  []byte
		resp StatusMessage
		err  error
		code int
	)
	if raw, err = json.Marshal(a); err != nil {
		return StatusMessage{}, errors.Wrap(err, "marshal request")
	}
	if raw, code, err = r.post(ctx, "api/snapshots", nil, raw); err != nil {
		return StatusMessage{}, errors.Wrap(err, "create snapshot")
	}
	if code/100 != 2 {
		return StatusMessage{}, fmt.Errorf("bad response: %d", code)
	}
	if err = json.Unmarshal(raw, &resp); err != nil {
		return StatusMessage{}, errors.Wrap(err, "unmarshal response message")
	}
	return resp, nil
}
