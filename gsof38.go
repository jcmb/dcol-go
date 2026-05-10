package dcol

import (
	"encoding/binary"
	"math"
)

// GSOFMessagePositionTypeInformation is Trimble GSOF output record type 38
// (Position Type Information). Payload layout is big-endian and has grown across
// receiver firmware; see the MinLen constants below.
const GSOFMessagePositionTypeInformation = 38

// Minimum inner payload lengths for GSOF type 38 (bytes required to reach the
// end of the last field in that firmware band). Layout matches Trimble docs /
// trimble_gsof_msgs PositionTypeInformation38.msg.
const (
	// Through NetworkFlags (Char): error scale + solution + RTK + correction age + net1.
	GSOFPositionTypeInformationMinLenThroughNetworkFlags = 11 // fw before 4.82 additions
	// Adds NetworkFlags2, FrameFlag, ITRFEpoch, TectonicPlate (fw 4.82+).
	GSOFPositionTypeInformationMinLenThroughTectonicPlate = 16
	// Adds RTXRAMSubMinutesLeft (int32) and PoleWobbleStatusFlag (fw 4.90+).
	GSOFPositionTypeInformationMinLenThroughPoleWobbleStatus = 21
	// Adds PoleWobbleDistance (float32) and PositionFixType (fw 4.94+).
	GSOFPositionTypeInformationMinLenFullKnown = 26
)

// GSOFPositionTypeInformation is the decoded body of a GSOF type-38 sub-record
// (bytes inside the [type][len][payload] envelope, i.e. the len-sized payload only).
// Fields beyond the end of the slice are left at zero values; DecodeGSOFPositionTypeInformation
// reports how many bytes were consumed from the known layout (not counting unknown future tail).
type GSOFPositionTypeInformation struct {
	ErrorScale           float32 // fw 4.40+ (replaces earlier 4 reserved bytes)
	SolutionFlags        uint8
	RTKCondition         uint8
	CorrectionAge        float32
	NetworkFlags         uint8
	NetworkFlags2        uint8 // fw 4.82+
	FrameFlag            uint8 // fw 4.82+
	ITRFEpoch            int16 // 1/100 year since 2005-01-01; fw 4.82+
	TectonicPlate        uint8 // fw 4.82+
	RTXRAMSubMinutesLeft int32 // fw 4.90+
	PoleWobbleStatusFlag uint8 // fw 4.90+
	PoleWobbleDistance   float32 // fw 4.94+
	PositionFixType      uint8 // fw 4.94+
}

// DecodeGSOFPositionTypeInformation decodes the big-endian type-38 payload.
// It never returns an error: short buffers stop at the first incomplete field;
// trailing bytes beyond the known layout are ignored for decoding but reflected
// only in the returned n when the full known layout is present (n is capped at
// GSOFPositionTypeInformationMinLenFullKnown).
func DecodeGSOFPositionTypeInformation(payload []byte) (GSOFPositionTypeInformation, int) {
	var out GSOFPositionTypeInformation
	b := payload
	off := 0
	need := func(n int) bool { return off+n <= len(b) }

	if !need(4) {
		return out, off
	}
	out.ErrorScale = math.Float32frombits(binary.BigEndian.Uint32(b[off:]))
	off += 4

	if !need(1) {
		return out, off
	}
	out.SolutionFlags = b[off]
	off++

	if !need(1) {
		return out, off
	}
	out.RTKCondition = b[off]
	off++

	if !need(4) {
		return out, off
	}
	out.CorrectionAge = math.Float32frombits(binary.BigEndian.Uint32(b[off:]))
	off += 4

	if !need(1) {
		return out, off
	}
	out.NetworkFlags = b[off]
	off++

	if !need(1) {
		return out, off
	}
	out.NetworkFlags2 = b[off]
	off++

	if !need(1) {
		return out, off
	}
	out.FrameFlag = b[off]
	off++

	if !need(2) {
		return out, off
	}
	out.ITRFEpoch = int16(binary.BigEndian.Uint16(b[off:]))
	off += 2

	if !need(1) {
		return out, off
	}
	out.TectonicPlate = b[off]
	off++

	if !need(4) {
		return out, off
	}
	out.RTXRAMSubMinutesLeft = int32(binary.BigEndian.Uint32(b[off:]))
	off += 4

	if !need(1) {
		return out, off
	}
	out.PoleWobbleStatusFlag = b[off]
	off++

	if !need(4) {
		return out, off
	}
	out.PoleWobbleDistance = math.Float32frombits(binary.BigEndian.Uint32(b[off:]))
	off += 4

	if !need(1) {
		return out, off
	}
	out.PositionFixType = b[off]
	off++

	return out, off
}
