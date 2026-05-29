package service

func validPNGBytes() []byte {
	return []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0x00, 0x00, 0x00, 0x0d}
}

func validPDFBytes() []byte {
	return []byte("%PDF-1.4\n%test\n")
}
