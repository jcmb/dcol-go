package dcol

import (
	"encoding/binary"
	"math"
	"testing"
)

func be32(f float32) uint32 { return math.Float32bits(f) }

func buildGSOF38Payload(t *testing.T) []byte {
	t.Helper()
	var p []byte
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, be32(1.5))
	p = append(p, buf...) // error scale
	p = append(p, 0x10, 0x20)
	binary.BigEndian.PutUint32(buf, be32(3.25))
	p = append(p, buf...) // correction age
	p = append(p, 0x30, 0x31) // network, network2
	p = append(p, 0x40) // frame
	binary.BigEndian.PutUint16(buf[:2], 0x0506)
	p = append(p, buf[:2]...) // ITRF epoch
	p = append(p, 0x07) // tectonic
	binary.BigEndian.PutUint32(buf, 0x08090A0B)
	p = append(p, buf...) // RTX minutes
	p = append(p, 0x0C) // pole status
	binary.BigEndian.PutUint32(buf, be32(9.0))
	p = append(p, buf...) // pole distance
	p = append(p, 0x0D) // position fix type
	if len(p) != GSOFPositionTypeInformationMinLenFullKnown {
		t.Fatalf("golden len %d want %d", len(p), GSOFPositionTypeInformationMinLenFullKnown)
	}
	return p
}

func TestDecodeGSOFPositionTypeInformationFull(t *testing.T) {
	p := buildGSOF38Payload(t)
	got, n := DecodeGSOFPositionTypeInformation(p)
	if n != len(p) {
		t.Fatalf("n=%d want %d", n, len(p))
	}
	if got.ErrorScale != 1.5 || got.CorrectionAge != 3.25 || got.PoleWobbleDistance != 9.0 {
		t.Fatalf("floats %+v", got)
	}
	if got.SolutionFlags != 0x10 || got.RTKCondition != 0x20 {
		t.Fatalf("flags %+v", got)
	}
	if got.NetworkFlags != 0x30 || got.NetworkFlags2 != 0x31 || got.FrameFlag != 0x40 {
		t.Fatalf("net/frame %+v", got)
	}
	if got.ITRFEpoch != int16(0x0506) || got.TectonicPlate != 0x07 {
		t.Fatalf("epoch/plate %+v", got)
	}
	if got.RTXRAMSubMinutesLeft != int32(0x08090A0B) {
		t.Fatalf("rtx %d", got.RTXRAMSubMinutesLeft)
	}
	if got.PoleWobbleStatusFlag != 0x0C || got.PositionFixType != 0x0D {
		t.Fatalf("tail %+v", got)
	}
}

func TestDecodeGSOFPositionTypeInformationPartial(t *testing.T) {
	full := buildGSOF38Payload(t)
	_, n11 := DecodeGSOFPositionTypeInformation(full[:GSOFPositionTypeInformationMinLenThroughNetworkFlags])
	if n11 != GSOFPositionTypeInformationMinLenThroughNetworkFlags {
		t.Fatalf("n11=%d", n11)
	}
	_, n16 := DecodeGSOFPositionTypeInformation(full[:GSOFPositionTypeInformationMinLenThroughTectonicPlate])
	if n16 != GSOFPositionTypeInformationMinLenThroughTectonicPlate {
		t.Fatalf("n16=%d", n16)
	}
	_, n21 := DecodeGSOFPositionTypeInformation(full[:GSOFPositionTypeInformationMinLenThroughPoleWobbleStatus])
	if n21 != GSOFPositionTypeInformationMinLenThroughPoleWobbleStatus {
		t.Fatalf("n21=%d", n21)
	}
	got3, n3 := DecodeGSOFPositionTypeInformation(full[:3])
	if n3 != 0 || got3.ErrorScale != 0 {
		t.Fatalf("short prefix should decode nothing: n=%d v=%+v", n3, got3)
	}
}

func TestDecodeGSOFPositionTypeInformationIgnoresFutureTail(t *testing.T) {
	full := buildGSOF38Payload(t)
	extra := append(append([]byte(nil), full...), 0xAA, 0xBB, 0xCC)
	got, n := DecodeGSOFPositionTypeInformation(extra)
	if n != GSOFPositionTypeInformationMinLenFullKnown {
		t.Fatalf("n=%d", n)
	}
	gotFull, _ := DecodeGSOFPositionTypeInformation(full)
	if got != gotFull {
		t.Fatalf("with tail differs: %+v vs %+v", got, gotFull)
	}
}

func TestExpandedToFlatBufferChunksOversizedInner(t *testing.T) {
	// expandedToFlatBuffer must not truncate inners >255 (defensive / forward-compatible).
	inner := make([]byte, 300)
	for i := range inner {
		inner[i] = byte(i)
	}
	flat := expandedToFlatBuffer([]ExpandedRecord{{MsgType: GSOFMessagePositionTypeInformation, Inner: inner}})
	if len(flat) != 2+255+2+45 {
		t.Fatalf("flat len %d", len(flat))
	}
	var merged []byte
	ptr := 0
	for ptr+2 <= len(flat) {
		n := int(flat[ptr+1])
		merged = append(merged, flat[ptr+2:ptr+2+n]...)
		ptr += 2 + n
	}
	if len(merged) != 300 {
		t.Fatalf("merged len %d", len(merged))
	}
	for i := range merged {
		if merged[i] != byte(i) {
			t.Fatalf("byte %d", i)
		}
	}
}
