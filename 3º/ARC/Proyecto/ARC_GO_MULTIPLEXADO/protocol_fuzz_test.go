package protocol

import (
	"bytes"
	"math/bits"
	"testing"
)

func FuzzDecode(f *testing.F) {
	for typ := REGISTER; typ <= ERROR; typ++ {
		p := fixturePacket(typ)
		f.Add(wire(f, p))
		if typ == POSITION || typ == RETRY_POSITION || typ == ACK {
			p.Flags = ServerToClient
			if typ != ACK {
				p.TargetID = 2
			}
			f.Add(wire(f, p))
		}
	}
	for _, v := range []uint32{1, 32} {
		p := fixturePacket(REGISTER_ACK)
		p.Registration.V = v
		f.Add(wire(f, p))
	}
	f.Add([]byte{})
	f.Add([]byte("ARCG"))
	f.Add(make([]byte, MaxPacketSize+1))
	manifests := [3]*Manifest{fixtureManifest(f, 1), fixtureManifest(f, 25), fixtureManifest(f, 32)}
	f.Fuzz(func(t *testing.T, src []byte) {
		p, err := Decode(src)
		if err != nil {
			if p != (Packet{}) {
				t.Fatal("error exposed partially decoded packet")
			}
			return
		}
		var dst [MaxPacketSize]byte
		n, err := Encode(dst[:], &p)
		if err != nil {
			t.Fatalf("decoded packet cannot be encoded: %v", err)
		}
		if !bytes.Equal(src, dst[:n]) {
			t.Fatal("noncanonical roundtrip")
		}
		for _, m := range manifests {
			if err := m.Validate(&p); err == nil && p.Type == RETRY_POSITION {
				expected, err := ExpectedMask(p.OriginID, m.Config().V)
				if err != nil || p.MissingMask == 0 || p.MissingMask & ^expected != 0 {
					t.Fatal("invalid selective retry accepted")
				}
			}
		}
	})
}

func FuzzMasks(f *testing.F) {
	f.Add(uint32(1), uint32(1), uint32(0))
	f.Add(uint32(25), uint32(25), uint32(0x00ffffff))
	f.Add(uint32(32), uint32(32), uint32(0x7fffffff))
	f.Add(uint32(0), uint32(0), uint32(0xffffffff))
	f.Fuzz(func(t *testing.T, id, v, received uint32) {
		expected, err := ExpectedMask(id, v)
		if id == 0 || v == 0 || v > 32 {
			if err == nil {
				t.Fatal("invalid mask parameters accepted")
			}
			return
		}
		if err != nil || bits.OnesCount32(expected) != int(v)-1 {
			t.Fatal("incorrect neighbor count")
		}
		// Independent bit-by-bit reference for every slot, including bit 31.
		var reference uint32
		for slot := uint32(0); slot < v; slot++ {
			if slot != (id-1)%v {
				reference |= uint32(1) << slot
			}
		}
		if expected != reference {
			t.Fatal("mask differs from reference")
		}
		missing, err := MissingMask(expected, received)
		if received & ^expected != 0 {
			if err == nil {
				t.Fatal("accepted unexpected ACK bits")
			}
			return
		}
		if err != nil || missing&received != 0 || missing|received != expected {
			t.Fatal("incorrect mask partition")
		}
	})
}
