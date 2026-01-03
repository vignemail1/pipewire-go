// Package spa - audio.go
// Audio-specific SPA/POD type definitions

package spa

// AudioFormat represents audio format specifications in SPA
type AudioFormat uint32

// Standard audio formats supported by PipeWire
const (
	AudioFormatUnknown   AudioFormat = 0
	AudioFormatS8        AudioFormat = 1  // Signed 8-bit
	AudioFormatU8        AudioFormat = 2  // Unsigned 8-bit
	AudioFormatS16LE     AudioFormat = 3  // Signed 16-bit Little-Endian
	AudioFormatS16BE     AudioFormat = 4  // Signed 16-bit Big-Endian
	AudioFormatU16LE     AudioFormat = 5  // Unsigned 16-bit Little-Endian
	AudioFormatU16BE     AudioFormat = 6  // Unsigned 16-bit Big-Endian
	AudioFormatS24LE     AudioFormat = 7  // Signed 24-bit Little-Endian (packed)
	AudioFormatS24BE     AudioFormat = 8  // Signed 24-bit Big-Endian (packed)
	AudioFormatU24LE     AudioFormat = 9  // Unsigned 24-bit Little-Endian (packed)
	AudioFormatU24BE     AudioFormat = 10 // Unsigned 24-bit Big-Endian (packed)
	AudioFormatS32LE     AudioFormat = 11 // Signed 32-bit Little-Endian
	AudioFormatS32BE     AudioFormat = 12 // Signed 32-bit Big-Endian
	AudioFormatU32LE     AudioFormat = 13 // Unsigned 32-bit Little-Endian
	AudioFormatU32BE     AudioFormat = 14 // Unsigned 32-bit Big-Endian
	AudioFormatS24_32LE  AudioFormat = 15 // Signed 24-bit in 32-bit LE
	AudioFormatS24_32BE  AudioFormat = 16 // Signed 24-bit in 32-bit BE
	AudioFormatU24_32LE  AudioFormat = 17 // Unsigned 24-bit in 32-bit LE
	AudioFormatU24_32BE  AudioFormat = 18 // Unsigned 24-bit in 32-bit BE
	AudioFormatF32LE     AudioFormat = 19 // Float 32-bit Little-Endian
	AudioFormatF32BE     AudioFormat = 20 // Float 32-bit Big-Endian
	AudioFormatF64LE     AudioFormat = 21 // Float 64-bit Little-Endian
	AudioFormatF64BE     AudioFormat = 22 // Float 64-bit Big-Endian
	AudioFormatS16OE     AudioFormat = 23 // Signed 16-bit Other-Endian
	AudioFormatU16OE     AudioFormat = 24 // Unsigned 16-bit Other-Endian
	AudioFormatS24OE     AudioFormat = 25 // Signed 24-bit Other-Endian
	AudioFormatU24OE     AudioFormat = 26 // Unsigned 24-bit Other-Endian
	AudioFormatS32OE     AudioFormat = 27 // Signed 32-bit Other-Endian
	AudioFormatU32OE     AudioFormat = 28 // Unsigned 32-bit Other-Endian
	AudioFormatF32OE     AudioFormat = 29 // Float 32-bit Other-Endian
	AudioFormatF64OE     AudioFormat = 30 // Float 64-bit Other-Endian
	AudioFormatS24_32OE  AudioFormat = 31 // Signed 24-bit in 32-bit OE
	AudioFormatU24_32OE  AudioFormat = 32 // Unsigned 24-bit in 32-bit OE
)

// String returns the audio format as a string
func (af AudioFormat) String() string {
	switch af {
	case AudioFormatUnknown:
		return "Unknown"
	case AudioFormatS8:
		return "S8"
	case AudioFormatU8:
		return "U8"
	case AudioFormatS16LE:
		return "S16LE"
	case AudioFormatS16BE:
		return "S16BE"
	case AudioFormatU16LE:
		return "U16LE"
	case AudioFormatU16BE:
		return "U16BE"
	case AudioFormatS24LE:
		return "S24LE"
	case AudioFormatS24BE:
		return "S24BE"
	case AudioFormatU24LE:
		return "U24LE"
	case AudioFormatU24BE:
		return "U24BE"
	case AudioFormatS32LE:
		return "S32LE"
	case AudioFormatS32BE:
		return "S32BE"
	case AudioFormatU32LE:
		return "U32LE"
	case AudioFormatU32BE:
		return "U32BE"
	case AudioFormatS24_32LE:
		return "S24_32LE"
	case AudioFormatS24_32BE:
		return "S24_32BE"
	case AudioFormatU24_32LE:
		return "U24_32LE"
	case AudioFormatU24_32BE:
		return "U24_32BE"
	case AudioFormatF32LE:
		return "F32LE"
	case AudioFormatF32BE:
		return "F32BE"
	case AudioFormatF64LE:
		return "F64LE"
	case AudioFormatF64BE:
		return "F64BE"
	default:
		return "Unknown"
	}
}

// AudioChannelPosition represents positions of audio channels
type AudioChannelPosition uint32

// Standard audio channel positions
const (
	AudioChannelPositionMono AudioChannelPosition = 0

	// Stereo positions
	AudioChannelPositionFL AudioChannelPosition = 1 // Front Left
	AudioChannelPositionFR AudioChannelPosition = 2 // Front Right

	// 5.1 positions
	AudioChannelPositionFC  AudioChannelPosition = 3 // Front Center
	AudioChannelPositionLFE AudioChannelPosition = 4 // Low-Frequency Effects
	AudioChannelPositionBL  AudioChannelPosition = 5 // Back Left
	AudioChannelPositionBR  AudioChannelPosition = 6 // Back Right

	// 7.1 positions
	AudioChannelPositionFLC AudioChannelPosition = 7 // Front Left Center
	AudioChannelPositionFRC AudioChannelPosition = 8 // Front Right Center
	AudioChannelPositionBC  AudioChannelPosition = 9 // Back Center
	AudioChannelPositionSL  AudioChannelPosition = 10 // Side Left
	AudioChannelPositionSR  AudioChannelPosition = 11 // Side Right

	// Top positions
	AudioChannelPositionTC  AudioChannelPosition = 12 // Top Center
	AudioChannelPositionTFL AudioChannelPosition = 13 // Top Front Left
	AudioChannelPositionTFR AudioChannelPosition = 14 // Top Front Right
	AudioChannelPositionTFC AudioChannelPosition = 15 // Top Front Center
	AudioChannelPositionTBL AudioChannelPosition = 16 // Top Back Left
	AudioChannelPositionTBR AudioChannelPosition = 17 // Top Back Right
	AudioChannelPositionTBC AudioChannelPosition = 18 // Top Back Center
	AudioChannelPositionBLC AudioChannelPosition = 19 // Back Left Center
	AudioChannelPositionBRC AudioChannelPosition = 20 // Back Right Center
	AudioChannelPositionLW  AudioChannelPosition = 21 // Left Wide
	AudioChannelPositionRW  AudioChannelPosition = 22 // Right Wide
)

// String returns the channel position as a string
func (acp AudioChannelPosition) String() string {
	switch acp {
	case AudioChannelPositionMono:
		return "Mono"
	case AudioChannelPositionFL:
		return "FL"
	case AudioChannelPositionFR:
		return "FR"
	case AudioChannelPositionFC:
		return "FC"
	case AudioChannelPositionLFE:
		return "LFE"
	case AudioChannelPositionBL:
		return "BL"
	case AudioChannelPositionBR:
		return "BR"
	case AudioChannelPositionFLC:
		return "FLC"
	case AudioChannelPositionFRC:
		return "FRC"
	case AudioChannelPositionBC:
		return "BC"
	case AudioChannelPositionSL:
		return "SL"
	case AudioChannelPositionSR:
		return "SR"
	case AudioChannelPositionTC:
		return "TC"
	case AudioChannelPositionTFL:
		return "TFL"
	case AudioChannelPositionTFR:
		return "TFR"
	case AudioChannelPositionTFC:
		return "TFC"
	case AudioChannelPositionTBL:
		return "TBL"
	case AudioChannelPositionTBR:
		return "TBR"
	case AudioChannelPositionTBC:
		return "TBC"
	case AudioChannelPositionBLC:
		return "BLC"
	case AudioChannelPositionBRC:
		return "BRC"
	case AudioChannelPositionLW:
		return "LW"
	case AudioChannelPositionRW:
		return "RW"
	default:
		return "Unknown"
	}
}

// AudioInfo represents complete audio format information
type AudioInfo struct {
	Format      AudioFormat
	Rate        uint32                // Sample rate (44100, 48000, etc.)
	Channels    uint32                // Number of channels
	Positions   []AudioChannelPosition // Channel positions
	ChannelMask uint64                // Bitmask of channels
}

// AudioProperties represents audio-related properties
type AudioProperties struct {
	Format     AudioFormat
	Rate       uint32
	Channels   uint32
	Quantum    uint32 // Default quantum size
	MaxLatency uint32 // Maximum latency in samples
}

// VideoFormat represents video format specifications
type VideoFormat uint32

// Standard video formats
const (
	VideoFormatUnknown VideoFormat = 0
	VideoFormatRGB     VideoFormat = 1
	VideoFormatBGR     VideoFormat = 2
	VideoFormatYUV     VideoFormat = 3
	VideoFormatNV12    VideoFormat = 4
	VideoFormatYUYV    VideoFormat = 5
	VideoFormatUYVY    VideoFormat = 6
	VideoFormatY41P    VideoFormat = 7
	VideoFormatIYUV    VideoFormat = 8
	VideoFormatI420    VideoFormat = 9
)

// String returns the video format as a string
func (vf VideoFormat) String() string {
	switch vf {
	case VideoFormatRGB:
		return "RGB"
	case VideoFormatBGR:
		return "BGR"
	case VideoFormatYUV:
		return "YUV"
	case VideoFormatNV12:
		return "NV12"
	case VideoFormatYUYV:
		return "YUYV"
	case VideoFormatUYVY:
		return "UYVY"
	default:
		return "Unknown"
	}
}

// MidiFormat represents MIDI format specifications
type MidiFormat uint32

// Standard MIDI formats
const (
	MidiFormatUnknown MidiFormat = 0
	MidiFormatCME     MidiFormat = 1 // Core MIDI Event
	MidiFormatSMPTE   MidiFormat = 2 // SMPTE timing
)

// String returns the MIDI format as a string
func (mf MidiFormat) String() string {
	switch mf {
	case MidiFormatCME:
		return "CME"
	case MidiFormatSMPTE:
		return "SMPTE"
	default:
		return "Unknown"
	}
}

// MediaType represents the type of media
type MediaType uint32

const (
	MediaTypeUnknown   MediaType = 0
	MediaTypeAudio     MediaType = 1
	MediaTypeVideo     MediaType = 2
	MediaTypeMidi      MediaType = 3
	MediaTypeControl   MediaType = 4
	MediaTypeApplication MediaType = 5
)

// String returns the media type as a string
func (mt MediaType) String() string {
	switch mt {
	case MediaTypeAudio:
		return "Audio"
	case MediaTypeVideo:
		return "Video"
	case MediaTypeMidi:
		return "Midi"
	case MediaTypeControl:
		return "Control"
	case MediaTypeApplication:
		return "Application"
	default:
		return "Unknown"
	}
}
