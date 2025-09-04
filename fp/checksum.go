package fp

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
	"log"
	"os"
	"strconv"
)

const SHA1LEN = 40
const SHA256LEN = 64

const (
	SHA1    = "SHA1"
	SHA256  = "SHA256"
	UNKNOWN = ""
)

type Checksum string

func NewChecksum(checksum string) Checksum {
	if ChecksumGetAlgo(checksum) == UNKNOWN {
		panic("Unknown checksum size")
	}
	return Checksum(checksum)
}

func ChecksumBlob(path string, algo string) Checksum {
	// TODO: os.Stat follows symlinks apparently
	s, err := os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}

	fileToRead, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer fileToRead.Close()

	var hashAlgo hash.Hash
	hashAlgo = checksumGoHash(algo)
	header := []byte("blob " + strconv.FormatInt(s.Size(), 10) + "\x00")
	hashAlgo.Write(header)
	if _, err := io.Copy(hashAlgo, fileToRead); err != nil {
		log.Fatal(err)
	}
	return Checksum(hex.EncodeToString(hashAlgo.Sum(nil)))
}

func checksumGoHash(algo string) hash.Hash {
	switch algo {
	case SHA1:
		return sha1.New()
	case SHA256:
		return sha256.New()
	}
	panic("tried to use algorithm " + algo)
}

func ChecksumGetAlgo(checksum string) string {
	if len(checksum) == SHA1LEN {
		return SHA1
	} else if len(checksum) == SHA256LEN {
		return SHA256
	} else {
		return UNKNOWN
	}
}

func (c Checksum) GetAlgo() string {
	return ChecksumGetAlgo(string(c))
}
