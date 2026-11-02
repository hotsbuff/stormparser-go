package mpq

var cryptTable = makeCryptTable()

func makeCryptTable() *[0x500]uint32 {
	var table [0x500]uint32
	seed := uint32(0x00100001)
	for index1 := 0; index1 < 0x100; index1++ {
		for i := 0; i < 5; i++ {
			index2 := i*0x100 + index1
			seed = (seed*125 + 3) % 0x2aaaab
			high := (seed & 0xffff) << 16
			seed = (seed*125 + 3) % 0x2aaaab
			low := seed & 0xffff
			table[index2] = high | low
		}
	}
	return &table
}

func HashString(value string, hashType HashType) uint32 {
	seed1 := uint32(0x7fed7fed)
	seed2 := uint32(0xeeeeeeee)
	for _, r := range []byte(value) {
		if r == '/' {
			r = '\\'
		}
		if r >= 'a' && r <= 'z' {
			r -= 'a' - 'A'
		}
		seed1 = cryptTable[uint32(hashType)<<8|uint32(r)] ^ (seed1 + seed2)
		seed2 = uint32(r) + seed1 + seed2 + (seed2 << 5) + 3
	}
	return seed1
}

func decrypt(data []byte, key uint32) []byte {
	result := make([]byte, len(data))
	seed1 := key
	seed2 := uint32(0xeeeeeeee)
	for i := 0; i+4 <= len(data); i += 4 {
		seed2 += cryptTable[0x400+(seed1&0xff)]
		value := uint32(data[i]) | uint32(data[i+1])<<8 | uint32(data[i+2])<<16 | uint32(data[i+3])<<24
		value ^= seed1 + seed2
		seed1 = ((^seed1 << 21) + 0x11111111) | (seed1 >> 11)
		seed2 = value + seed2 + (seed2 << 5) + 3
		result[i] = byte(value)
		result[i+1] = byte(value >> 8)
		result[i+2] = byte(value >> 16)
		result[i+3] = byte(value >> 24)
	}
	copy(result[len(result)-len(result)%4:], data[len(data)-len(data)%4:])
	return result
}
