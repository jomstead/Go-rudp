package packet

type Ack struct {
	Data uint32
}

type Packet struct {
	Seq       uint32
	Data      []byte
	Timestamp int64
}

func (a *Ack) Set(flag uint32) {
	flag = PowInts(2, flag)
	a.Data = a.Data | flag
}

func (a *Ack) Clear(flag uint32) {
	flag = PowInts(2, flag)
	a.Data = a.Data &^ flag
}

func (a Ack) Has(flag uint32) bool {
	flag = PowInts(2, flag)
	return a.Data&flag != 0
}

func (a *Ack) Shift(count uint32) {
	a.Data = a.Data << count
}

func UpdateAcknowledgements(seq uint32, remote uint32, bitfield *Ack) uint32 {
	// shift the remote_acks bitfield by the difference in new vs last sequence number received
	shift := int32(seq - remote)
	if remote != ^uint32(0) {
		// this is a packet received with a newer sequence number (received 'in order')
		if shift > 0 {
			// push the current sequence into the history bitfield
			bitfield.Shift(1)
			bitfield.Set(0)
			// shift further if there is a gap in the sequence numbers
			bitfield.Shift(uint32(shift) - 1)
		} else {
			// this is a packet received with an older sequence number (received 'out of order')
			// set the bitfield for that sequence number as received
			bitfield.Set(uint32(-shift) - 1)
		}
	}
	// update the remote sequence number if the new seq is larger or if the remote sequence is still the initial value
	if seq > remote || remote == ^uint32(0) {
		remote = seq
	}
	return remote
}

// Assumption: n >= 0
func PowInts(x, n uint32) uint32 {
	if n == 0 {
		return 1
	}
	if n == 1 {
		return x
	}
	y := PowInts(x, n/2)
	if n%2 == 0 {
		return y * y
	}
	return x * y * y
}
