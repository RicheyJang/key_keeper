package safer

import (
	"crypto/rand"
	"encoding/binary"

	"golang.org/x/crypto/sha3"
)

func getKeyContent(length uint, id uint, version uint, mainKey []byte, ss []byte) ([]byte, error) {
	hash := sha3.NewShake128()
	err := binary.Write(hash, binary.LittleEndian, uint64(id))
	if err != nil {
		return nil, err
	}
	if _, err = hash.Write(mainKey); err != nil {
		return nil, err
	}
	if err = binary.Write(hash, binary.LittleEndian, uint64(version)); err != nil {
		return nil, err
	}
	if _, err = hash.Write(ss); err != nil {
		return nil, err
	}
	res := make([]byte, length)
	_, _ = hash.Read(res)
	return res, nil
}

const ssLength = 32
const mainKeyLength = 32

func getNewSS(length int) ([]byte, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}
