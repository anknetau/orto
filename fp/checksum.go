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

type Checksum string
type Algo string

const (
	SHA1    Algo = "SHA1"
	SHA256  Algo = "SHA256"
	UNKNOWN Algo = ""
)

func NewChecksum(checksum string) Checksum {
	if AlgoOfGitHashValue(checksum) == UNKNOWN {
		panic("Unknown checksum size")
	}
	return Checksum(checksum)
}

func InternalChecksumBlob(path string, algo Algo) Checksum {
	// TODO: os.Stat follows symlinks apparently
	fileInfo, err := os.Stat(path)
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
	header := []byte("blob " + strconv.FormatInt(fileInfo.Size(), 10) + "\x00")
	hashAlgo.Write(header)
	if _, err := io.Copy(hashAlgo, fileToRead); err != nil {
		log.Fatal(err)
	}
	return Checksum(hex.EncodeToString(hashAlgo.Sum(nil)))
}

func checksumGoHash(algo Algo) hash.Hash {
	switch algo {
	case SHA1:
		return sha1.New()
	case SHA256:
		return sha256.New()
	}
	panic("tried to use algorithm " + algo)
}

func AlgoOfGitHashValue(checksum string) Algo {
	if len(checksum) == SHA1LEN {
		return SHA1
	} else if len(checksum) == SHA256LEN {
		return SHA256
	} else {
		return UNKNOWN
	}
}

func AlgoOfGitObjectFormat(format string) Algo {
	switch format {
	case "sha1":
		return SHA1
	case "sha256":
		return SHA256
	default:
		return UNKNOWN
	}
}

func (c Checksum) GetAlgo() Algo {
	return AlgoOfGitHashValue(string(c))
}
