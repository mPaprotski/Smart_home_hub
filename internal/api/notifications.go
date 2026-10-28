package api

import (
	"fmt"
	"github.com/mPaprotski/Smart_home_hub/internal/events"
	"time"
)

// A bounded, non-sensitive notification follows SQL commit. SQL audit is authoritative.
func (s *Server) notifyAudit() {
	now := time.Now().UTC()
	seq := s.notificationSequence.Add(1)
	if err := s.bus.Publish(events.Envelope{ID: fmt.Sprintf("%s-%d", s.notificationEpoch, seq), Version: 1, Type: "SYSTEM_STATE_CHANGED", Producer: "core.api", OccurredAt: now, ReceivedAt: now, BootEpoch: s.notificationEpoch, Sequence: seq, Resource: events.Resource{Kind: "system", ID: "security"}, Revision: seq, Payload: []byte(`{"kind":"security_audit_changed"}`)}); err != nil {
		s.notificationFailures.Add(1)
	}
}
