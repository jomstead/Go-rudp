package packet

import (
	"log"
	"testing"
)

func TestRUDP_AckSet(t *testing.T) {
	a := Ack{Data: 0}
	a.Set(2)
	log.Printf("%b", a.Data)
	if a.Data != 0b100 {
		t.Error("Ack set not working")
	}
}

func TestRUDP_AckClear(t *testing.T) {
	a := Ack{Data: 0b111}
	a.Clear(1)
	log.Printf("%b", a.Data)
	if a.Data != 0b101 {
		t.Error("Ack clear not working")
	}
}

func TestRUDP_AckShift(t *testing.T) {
	a := Ack{Data: 0b111}
	a.Shift(1)
	log.Printf("%b", a.Data)
	if a.Data != 0b1110 {
		t.Error("Ack shift not working")
	}
}

func TestRUDP_AckHas(t *testing.T) {
	a := Ack{Data: 0b100}
	if !a.Has(2) {
		t.Error("Ack has not working")
	}
}

func TestRUDP_UpdateAcknowledgements5320(t *testing.T) {
	remote_seq := ^uint32(0)
	bitfield := Ack{Data: 0}
	remote_seq = UpdateAcknowledgements(5, remote_seq, &bitfield)
	remote_seq = UpdateAcknowledgements(3, remote_seq, &bitfield)
	remote_seq = UpdateAcknowledgements(2, remote_seq, &bitfield)
	remote_seq = UpdateAcknowledgements(0, remote_seq, &bitfield)
	if bitfield.Data != 0b10110 {
		t.Errorf("Bitfield error, expected %b, received %b", 0b10110, bitfield.Data)
	}
	if remote_seq != 5 {
		t.Errorf("remote sequence error, expected %d, received %d", 5, remote_seq)
	}
}

func TestRUDP_UpdateAcknowledgementsSeq0(t *testing.T) {
	remote_seq := ^uint32(0)
	bitfield := Ack{Data: 0}
	remote_seq = UpdateAcknowledgements(0, remote_seq, &bitfield)
	if bitfield.Data != 0 {
		t.Errorf("Bitfield error, expected %b, received %b", 0, bitfield.Data)
	}
	if remote_seq != 0 {
		t.Errorf("remote sequence error, expected %d, received %d", 0, remote_seq)
	}
}

func TestRUDP_UpdateAcknowledgementsSeq5(t *testing.T) {
	remote_seq := ^uint32(0)
	bitfield := Ack{Data: 0}
	remote_seq = UpdateAcknowledgements(5, remote_seq, &bitfield)
	if bitfield.Data != 0 {
		t.Errorf("Bitfield error, expected %b, received %b", 0, bitfield.Data)
	}
	if remote_seq != 5 {
		t.Errorf("remote sequence error, expected %d, received %d", 5, remote_seq)
	}
}

func TestRUDP_UpdateAcknowledgementsSeq012(t *testing.T) {
	remote_seq := ^uint32(0)
	bitfield := Ack{Data: 0}
	remote_seq = UpdateAcknowledgements(0, remote_seq, &bitfield)
	remote_seq = UpdateAcknowledgements(1, remote_seq, &bitfield)
	remote_seq = UpdateAcknowledgements(2, remote_seq, &bitfield)
	if bitfield.Data != 0b11 {
		t.Errorf("Bitfield error, expected %b, received %b", 0b11, bitfield.Data)
	}
	if remote_seq != 2 {
		t.Errorf("remote sequence error, expected %d, received %d", 2, remote_seq)
	}
}

func TestRUDP_PowInt(t *testing.T) {
	if PowInts(2, 5) != 32 {
		t.Error("PowInts failed 2^5=32")
	}
	if PowInts(2, 6) != 64 {
		t.Error("PowInts failed 2^6=64")
	}
	if PowInts(2, 7) != 128 {
		t.Error("PowInts failed 2^7=128")
	}
}
