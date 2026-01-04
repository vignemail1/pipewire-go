// Package spa - Audio-specific types and constants
// spa/audio.go
// Complete audio format definitions, channel enums, and utility functions

package spa

import (
	"fmt"
)

// ===== Channel Enum =====

type Channel int

const (
	ChannelMono Channel = iota
	ChannelStereoLeft
	ChannelStereoRight
	ChannelFC  // Front Center
	ChannelLFE // Low Frequency Effects (Subwoofer)
	ChannelBL  // Back Left
	ChannelBR  // Back Right
	ChannelFLC // Front Left Center
	ChannelFRC // Front Right Center
	ChannelBC  // Back Center
	ChannelSL  // Side Left
	ChannelSR  // Side Right
	ChannelTC  // Top Center
	ChannelTFL // Top Front Left
	ChannelTFC // Top Front Center
	ChannelTFR // Top Front Right
	ChannelTBL // Top Back Left
	ChannelTBC // Top Back Center
	ChannelTBR // Top Back Right
)

func (c Channel) String() string {
	switch c {
	case ChannelMono:
		return "mono"
	case ChannelStereoLeft:
		return "stereo_left"
	case ChannelStereoRight:
		return "stereo_right"
	case ChannelFC:
		return "fc"
	case ChannelLFE:
		return "lfe"
	case ChannelBL:
		return "bl"
	case ChannelBR:
		return "br"
	case ChannelFLC:
		return "flc"
	case ChannelFRC:
		return "frc"
	case ChannelBC:
		return "bc"
	case ChannelSL:
		return "sl"
	case ChannelSR:
		return "sr"
	case ChannelTC:
		return "tc"
	case ChannelTFL:
		return "tfl"
	case ChannelTFC:
		return "tfc"
	case ChannelTFR:
		return "tfr"
	case ChannelTBL:
		return "tbl"
	case ChannelTBC:
		return "tbc"
	case ChannelTBR:
		return "tbr"
	default:
		return fmt.Sprintf("unknown(%d)", c)
	}
}

// ===== Audio Format Enum =====

type AudioFormat int

const (
	AudioFormatUnknown AudioFormat = iota
	AudioFormatF32                 // 32-bit floating point
	AudioFormatS16                 // 16-bit signed integer
	AudioFormatS24                 // 24-bit signed integer
	AudioFormatS32                 // 32-bit signed integer
	AudioFormatU32                 // 32-bit unsigned integer
	AudioFormatF64                 // 64-bit floating point
)

func (af AudioFormat) String() string {
	switch af {
	case AudioFormatF32:
		return "f32"
	case AudioFormatS16:
		return "s16"
	case AudioFormatS24:
		return "s24"
	case AudioFormatS32:
		return "s32"
	case AudioFormatU32:
		return "u32"
	case AudioFormatF64:
		return "f64"
	default:
		return "unknown"
	}
}

// SampleSize returns the size in bytes of a single sample
func (af AudioFormat) SampleSize() int {
	switch af {
	case AudioFormatS16:
		return 2
	case AudioFormatS24:
		return 3
	case AudioFormatS32, AudioFormatU32, AudioFormatF32:
		return 4
	case AudioFormatF64:
		return 8
	default:
		return 0
	}
}

// ===== Audio Properties =====

type AudioProperties struct {
	SampleRate  uint32
	Channels    uint32
	Format      AudioFormat
	ChannelMap  []Channel
	Interleaved bool
}

func NewAudioProperties() *AudioProperties {
	return &AudioProperties{
		SampleRate:  48000,
		Channels:    2,
		Format:      AudioFormatF32,
		Interleaved: true,
	}
}

// FrameSize returns the size of one audio frame in bytes
func (ap *AudioProperties) FrameSize() uint32 {
	return uint32(ap.Format.SampleSize()) * ap.Channels
}

// ===== Helper Functions =====

// AudioFormatFromString converts a string to AudioFormat
func AudioFormatFromString(s string) (AudioFormat, error) {
	switch s {
	case "f32":
		return AudioFormatF32, nil
	case "s16":
		return AudioFormatS16, nil
	case "s24":
		return AudioFormatS24, nil
	case "s32":
		return AudioFormatS32, nil
	case "u32":
		return AudioFormatU32, nil
	case "f64":
		return AudioFormatF64, nil
	default:
		return AudioFormatUnknown, fmt.Errorf("unknown audio format: %s", s)
	}
}

// ChannelFromString converts a string to Channel
func ChannelFromString(s string) (Channel, error) {
	switch s {
	case "mono":
		return ChannelMono, nil
	case "stereo_left":
		return ChannelStereoLeft, nil
	case "stereo_right":
		return ChannelStereoRight, nil
	case "fc":
		return ChannelFC, nil
	case "lfe":
		return ChannelLFE, nil
	case "bl":
		return ChannelBL, nil
	case "br":
		return ChannelBR, nil
	case "flc":
		return ChannelFLC, nil
	case "frc":
		return ChannelFRC, nil
	case "bc":
		return ChannelBC, nil
	case "sl":
		return ChannelSL, nil
	case "sr":
		return ChannelSR, nil
	case "tc":
		return ChannelTC, nil
	case "tfl":
		return ChannelTFL, nil
	case "tfc":
		return ChannelTFC, nil
	case "tfr":
		return ChannelTFR, nil
	case "tbl":
		return ChannelTBL, nil
	case "tbc":
		return ChannelTBC, nil
	case "tbr":
		return ChannelTBR, nil
	default:
		return ChannelMono, fmt.Errorf("unknown channel: %s", s)
	}
}

// GetChannelLayout returns standard channel layouts for common configurations
func GetChannelLayout(format AudioFormat, channels int) ([]Channel, error) {
	switch channels {
	case 1:
		return []Channel{ChannelMono}, nil
	case 2:
		return []Channel{ChannelStereoLeft, ChannelStereoRight}, nil
	case 4:
		return []Channel{ChannelFL, ChannelFR, ChannelBL, ChannelBR}, nil
	case 6:
		// 5.1 surround
		return []Channel{ChannelFL, ChannelFR, ChannelFC, ChannelLFE, ChannelBL, ChannelBR}, nil
	case 8:
		// 7.1 surround
		return []Channel{
			ChannelFL, ChannelFR, ChannelFC, ChannelLFE,
			ChannelBL, ChannelBR, ChannelSL, ChannelSR,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported channel count: %d", channels)
	}
}

// CalculateSampleSize returns the size of audio data in bytes
func CalculateSampleSize(format AudioFormat, channels uint32, samples uint32) uint32 {
	sampleSize := uint32(format.SampleSize())
	return sampleSize * channels * samples
}

// ===== Channel Aliases for Convenience =====

const (
	ChannelFL = ChannelStereoLeft
	ChannelFR = ChannelStereoRight
)

// ===== Audio Stream Configuration =====

type AudioStreamConfig struct {
	SampleRate uint32
	Channels   uint32
	Format     AudioFormat
	BufferSize uint32
	Periods    uint32
	ChannelMap []Channel
	Latency    uint32 // milliseconds
}

func NewAudioStreamConfig() *AudioStreamConfig {
	return &AudioStreamConfig{
		SampleRate: 48000,
		Channels:   2,
		Format:     AudioFormatF32,
		BufferSize: 256,
		Periods:    2,
		Latency:    10,
	}
}

// GetLatencySamples returns the latency in samples
func (asc *AudioStreamConfig) GetLatencySamples() uint32 {
	return (asc.SampleRate * asc.Latency) / 1000
}

// GetBitRate returns the bitrate in bits per second
func (asc *AudioStreamConfig) GetBitRate() uint32 {
	return asc.SampleRate * uint32(asc.Format.SampleSize()) * 8 * asc.Channels
}

// ===== Common Sample Rates =====

const (
	SampleRate8000   uint32 = 8000
	SampleRate16000  uint32 = 16000
	SampleRate22050  uint32 = 22050
	SampleRate24000  uint32 = 24000
	SampleRate44100  uint32 = 44100
	SampleRate48000  uint32 = 48000
	SampleRate88200  uint32 = 88200
	SampleRate96000  uint32 = 96000
	SampleRate176400 uint32 = 176400
	SampleRate192000 uint32 = 192000
)

// IsValidSampleRate checks if the sample rate is common
func IsValidSampleRate(sr uint32) bool {
	validRates := []uint32{
		SampleRate8000, SampleRate16000, SampleRate22050, SampleRate24000,
		SampleRate44100, SampleRate48000, SampleRate88200, SampleRate96000,
		SampleRate176400, SampleRate192000,
	}
	for _, rate := range validRates {
		if sr == rate {
			return true
		}
	}
	return false
}

// SampleRateString returns a string representation of the sample rate
func SampleRateString(sr uint32) string {
	switch sr {
	case SampleRate8000:
		return "8kHz"
	case SampleRate16000:
		return "16kHz"
	case SampleRate22050:
		return "22.05kHz"
	case SampleRate24000:
		return "24kHz"
	case SampleRate44100:
		return "44.1kHz"
	case SampleRate48000:
		return "48kHz"
	case SampleRate88200:
		return "88.2kHz"
	case SampleRate96000:
		return "96kHz"
	case SampleRate176400:
		return "176.4kHz"
	case SampleRate192000:
		return "192kHz"
	default:
		return fmt.Sprintf("%dHz", sr)
	}
}
