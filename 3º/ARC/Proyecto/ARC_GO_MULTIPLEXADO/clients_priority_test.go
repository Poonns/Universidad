package protocol

import (
	"fmt"
	"strings"
	"testing"
)

// These fixtures exercise the real selector, without sockets or RTO sleeps.
// Ticket zero is the existing synthetic-fixture convention. Refill immediately
// after every selection so queue emptiness cannot explain a scheduling choice.
func priorityFixture(t testing.TB, burst int, control, retry bool) *ClientGenerator {
	t.Helper()
	cfg := clientsConfig(t, 25, 25, 5, 1)
	cfg.HighBurst = burst
	g := bareClients(t, cfg)
	w := &g.workers[0]
	for _, q := range []struct {
		queue   *boundedQueue[clientOutbound]
		typ     MessageType
		enabled bool
	}{{w.high, ACK, true}, {w.control, METRICS, control}, {w.retry, RETRY_POSITION, retry}, {w.data, POSITION, true}} {
		if q.enabled && !q.queue.push(clientOutbound{Packet: Packet{Header: Header{Type: q.typ}}}, false) {
			t.Fatal("fixture enqueue")
		}
	}
	return g
}

func priorityStep(t testing.TB, g *ClientGenerator) byte {
	t.Helper()
	job, owner, ok, _ := g.takeOutbound(0)
	if !ok || owner != 0 {
		t.Fatal("eligible selector stalled", job, owner)
	}
	w := &g.workers[0]
	q, class := w.high, byte('H')
	switch job.Packet.Type {
	case METRICS:
		q, class = w.control, 'C'
	case RETRY_POSITION:
		q, class = w.retry, 'R'
	case POSITION:
		q, class = w.data, 'D'
	case ACK:
	default:
		t.Fatal("unexpected type", job.Packet.Type)
	}
	if !q.push(job, false) {
		t.Fatal("refill failed")
	}
	return class
}

func TestClientHighBurstRealSelectorRatioAndFairness(t *testing.T) {
	for _, burst := range []int{8, 24} {
		t.Run(fmt.Sprintf("HighBurst%d", burst), func(t *testing.T) {
			g := priorityFixture(t, burst, false, false)
			var got strings.Builder
			for i := 0; i < 3*(burst+1); i++ {
				got.WriteByte(priorityStep(t, g))
			}
			want := strings.Repeat(strings.Repeat("H", burst)+"D", 3)
			if got.String() != want || g.cfg.AckBurst != 4 {
				t.Fatal("HIGH/DATA ratio or unchanged AckBurst", got.String(), want)
			}
			t.Logf("HIGH+DATA continuously eligible: %s", got.String())
		})
	}
}

func TestClientHighBurstControlAndRetryPolicy(t *testing.T) {
	for _, burst := range []int{8, 24} {
		for _, retry := range []bool{false, true} {
			t.Run(fmt.Sprintf("burst%d/retry%t", burst, retry), func(t *testing.T) {
				g := priorityFixture(t, burst, true, retry)
				upper, counts := 0, map[byte]int{}
				var got strings.Builder
				// lcm(HighBurst, AckBurst+1) = 5*HighBurst for 8/24.
				for i := 0; i < 5*(burst+1); i++ {
					class := priorityStep(t, g)
					got.WriteByte(class)
					counts[class]++
					if class == 'D' || class == 'R' {
						if upper != burst {
							t.Fatal("lower turn did not follow exact upper quota", upper, class)
						}
						upper = 0
					} else {
						upper++
					}
				}
				if counts['H'] != 4*burst || counts['C'] != burst || upper != 0 {
					t.Fatal("CONTROL must share the upper quota with HIGH", counts, upper)
				}
				if retry {
					if counts['R'] != 5 || counts['D'] != 0 || g.workers[0].data.stats().Depth != 1 {
						t.Fatal("continuous retry backlog must defer DATA", counts)
					}
					// Once the finite recovery backlog drains, DATA must progress.
					g.workers[0].retry.pop()
					for i := 0; i < burst; i++ {
						priorityStep(t, g)
					}
					if priorityStep(t, g) != 'D' {
						t.Fatal("DATA did not resume after recovery")
					}
				} else if counts['D'] != 5 || counts['R'] != 0 {
					t.Fatal("DATA fairness lost", counts)
				}
				t.Logf("pattern=%s; HIGH=%d CONTROL=%d RETRY=%d DATA=%d", got.String(), counts['H'], counts['C'], counts['R'], counts['D'])
			})
		}
	}
}

func TestClientHighBurstIsNotCapWhenLowerIneligible(t *testing.T) {
	for _, burst := range []int{8, 24} {
		g := priorityFixture(t, burst, false, true)
		if !g.retries.take(0) {
			t.Fatal("fixture could not exhaust retry token")
		}
		for i := 0; i < 3*burst; i++ {
			if priorityStep(t, g) != 'H' {
				t.Fatal("ineligible lower class blocked HIGH")
			}
		}
		if g.highStreak != burst || g.workers[0].data.stats().Depth != 1 || g.workers[0].retry.stats().Depth != 1 {
			t.Fatal("quota saturation or retry backpressure changed")
		}
	}
}
