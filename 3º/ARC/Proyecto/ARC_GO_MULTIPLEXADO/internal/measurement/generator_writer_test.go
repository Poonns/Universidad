package measurement

import (
	arc "arc.local/multiplexado"
	"testing"
)

func TestGeneratorWriterPartitionFromFinalCounters(t *testing.T) {
	// Different counts exercise register/metrics retries with the original wire
	// type, selective POSITION retries, and reactive control ACKs in HIGH.
	st := arc.ClientStats{PacketsSent: 187, RegisterSent: 5, RegisterRetries: 2,
		MetricsSent: 4, MetricsRetries: 1, PositionSent: 7, PositionRetries: 3,
		AckSent: 168, StartAckSent: 7, GroupDoneAckSent: 3}
	st.PacketsSent = st.RegisterSent + st.MetricsSent + st.PositionSent + st.PositionRetries + st.AckSent + st.StartAckSent + st.GroupDoneAckSent
	w := GeneratorWriterSnapshot(st)
	if w.HighWrites != 178 || w.ControlWrites != 6 || w.RetryWrites != 6 || w.DataWrites != 7 || w.Writes != 197 ||
		w.Writes != w.HighWrites+w.ControlWrites+w.RetryWrites+w.DataWrites {
		t.Fatal("physical writes are not a disjoint class partition", w)
	}
}
