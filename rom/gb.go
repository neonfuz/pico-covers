package rom

func newGBROM(fileName string, header []byte, sha1Hex string) *ROM {
	return newROM(fileName, GB,
		WithHeader(header),
		WithSha1(sha1Hex),
	)
}
