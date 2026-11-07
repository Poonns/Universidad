package main

import "testing"

func TestShardPlanConservesAggregateBudgetsAndWholeGroups(t *testing.T) {
	p, err := shardPlan(30000, 20000, 128)
	if err != nil || len(p) != 2 {
		t.Fatal(p, err)
	}
	var count, workers, procs, position, register, retry, pw, rw int
	for i, g := range p {
		if g.GeneratorID != i+1 || g.FirstID != 1+i*15000 || g.Count != 15000 || (g.FirstID-1)%25 != 0 || g.Count%25 != 0 {
			t.Fatal(g)
		}
		count += g.Count
		workers += g.Workers
		procs += g.P
		position += g.PositionRate
		register += g.RegisterRate
		retry += g.RetryRate
		pw += g.PositionWindow
		rw += g.RegisterWindow
	}
	if count != 30000 || workers != 6 || procs != 6 || position != 20000 || register != 40000 || retry != 100 || pw != 128 || rw != 128 {
		t.Fatal(p)
	}
	for _, bad := range [][3]int{{30025, 20000, 128}, {30000, 20001, 128}, {30000, 20000, 127}, {25, 20000, 128}, {30000, 1, 128}, {30000, 20000, 1}} {
		if _, err := shardPlan(bad[0], bad[1], bad[2]); err == nil {
			t.Fatal("invalid split accepted", bad)
		}
	}
}
