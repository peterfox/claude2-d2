package r2

var (
	handshakeBytes = []byte{
		0x75, 0x73, 0x65, 0x74, 0x68, 0x65, 0x66, 0x6F,
		0x72, 0x63, 0x65, 0x2E, 0x2E, 0x2E, 0x62, 0x61,
		0x6E, 0x64,
	}

	Animations = map[string]byte{
		"charger_1":        0x00,
		"charger_2":        0x01,
		"charger_3":        0x02,
		"charger_4":        0x03,
		"charger_5":        0x04,
		"charger_6":        0x05,
		"charger_7":        0x06,
		"emote_alarm":      0x07,
		"emote_angry":      0x08,
		"emote_annoyed":    0x09,
		"emote_chatty":     0x0a,
		"emote_drive":      0x0b,
		"emote_excited":    0x0c,
		"emote_happy":      0x0d,
		"emote_ion_blast":  0x0e,
		"emote_laugh":      0x0f,
		"emote_no":         0x10,
		"emote_sad":        0x11,
		"emote_sassy":      0x12,
		"emote_scared":     0x13,
		"emote_spin":       0x14,
		"emote_yes":        0x15,
		"emote_scan":       0x16,
		"emote_sleep":      0x17,
		"emote_surprised":  0x18,
		"idle_1":           0x19,
		"idle_2":           0x1a,
		"idle_3":           0x1b,
		"patrol_alarm":     0x1c,
		"patrol_hit":       0x1d,
		"patrol_patrolling": 0x1e,
		"wwm_angry":        0x1f,
		"wwm_anxious":      0x20,
		"wwm_bow":          0x21,
		"wwm_concern":      0x22,
		"wwm_curious":      0x23,
		"wwm_double_take":  0x24,
		"wwm_excited":      0x25,
		"wwm_fiery":        0x26,
		"wwm_frustrated":   0x27,
		"wwm_happy":        0x28,
		"wwm_jittery":      0x29,
		"wwm_laugh":        0x2a,
		"wwm_long_shake":   0x2b,
		"wwm_no":           0x2c,
		"wwm_ominous":      0x2d,
		"wwm_relieved":     0x2e,
		"wwm_sad":          0x2f,
		"wwm_scared":       0x30,
		"wwm_shake":        0x31,
		"wwm_surprised":    0x32,
		"wwm_taunting":     0x33,
		"wwm_whisper":      0x34,
		"wwm_yelling":      0x35,
		"wwm_yoohoo":       0x36,
		"motor":            0x37,
	}
)

func initPacket() []byte {
	return buildPacket([]byte{0x0A, 0x13, 0x0D}, nil)
}

func animatePacket(animationID byte) []byte {
	return buildPacket([]byte{0x0A, 0x17, 0x05}, []byte{0x00, animationID})
}

func powerOffPacket() []byte {
	return buildPacket([]byte{0x0A, 0x13, 0x00}, nil)
}

func stopAnimationPacket() []byte {
	return buildPacket([]byte{0x0A, 0x17, 0x2B}, nil)
}
