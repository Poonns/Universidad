package measurement

import arc "arc.local/multiplexado"

type GeneratorWriterStats struct {
	Writes, HighWrites, ControlWrites, RetryWrites, DataWrites uint64
}

// Derive physical successful writes from existing counters after Close. No
// per-datagram increment, clock, lock or change to the generator's selector.
// REGISTER/METRICS retries use RETRY, despite sharing their original wire types.
func GeneratorWriterSnapshot(st arc.ClientStats) GeneratorWriterStats {
	return GeneratorWriterStats{
		Writes:        st.PacketsSent,
		HighWrites:    st.AckSent + st.StartAckSent + st.GroupDoneAckSent,
		ControlWrites: st.RegisterSent - st.RegisterRetries + st.MetricsSent - st.MetricsRetries,
		RetryWrites:   st.RegisterRetries + st.PositionRetries + st.MetricsRetries,
		DataWrites:    st.PositionSent,
	}
}
