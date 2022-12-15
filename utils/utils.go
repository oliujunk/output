package utils

func Crc16(data []byte, len int) uint16 {
	var crc uint16 = 0xFFFF
	var polynomial uint16 = 0xA001

	if len == 0 {
		return 0
	}

	for i := 0; i < len; i++ {
		crc ^= uint16(data[i]) & 0x00FF
		for j := 0; j < 8; j++ {
			if (crc & 0x0001) != 0 {
				crc >>= 1
				crc ^= polynomial
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}

func TestFrame(data []byte, len int) bool {
	crc := Crc16(data, len-2)
	testCrc := ((uint16)(data[len-1]) << 8) + (uint16)(data[len-2])
	if crc == testCrc {
		return true
	}
	return false
}
