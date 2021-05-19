package helper

import (
	"crypto/md5"
	"encoding/hex"
	vurl "net/url"
	"strings"

	"github.com/akrylysov/pogreb"
	"github.com/dgraph-io/badger/v3"
	"github.com/speps/go-hashids/v2"
)

// CreateURL
type CreateURL struct {
	URL string `json:"url"`
}

// DeleteURL - Delete
type DeleteURL struct {
	URL     string `json:"long"`
	ShortID string `json:"short"`
}

// UpdateURL - Delete
type UpdateURL struct {
	OURL    string `json:"old"`
	NURL    string `json:"new"`
	ShortID string `json:"short"`
}

// Ips
func IPs(ips []string) string {
	return strings.Join(ips, "-")
}

// validateURL is basic and not 100% correct validate URL, but this can be starting place...
func ValidateURL(url string) bool {
	if len(url) < 5 {
		return false
	}
	if _, err := vurl.ParseRequestURI(url); err != nil {
		return false
	}
	return true
}

// createMD5Hash md5 hash the URL
func CreateMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// generateToken generates uniqiue Token
func GenerateToken(salt string, num int) (string, error) {
	hd := hashids.NewData()
	hd.Salt = salt
	hd.Alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	hd.MinLength = ShortIDToken
	h, err := hashids.NewWithData(hd)
	if err != nil {
		return "", err
	}
	e, err := h.Encode([]int{num})
	if err != nil {
		return "", err
	}
	shortID := NodeID + string(salt[0:AddFromToken]) + e
	return shortID, nil
}

// errorPrint based on development or production server
func ErrorPrint(devError string, prodError string) string {
	if Production == 1 {
		return prodError
	}
	return devError
}

// findDB is to get the key & value from badger database
func FindDB(key []byte, db *pogreb.DB) (string, error) {
	// get the value from pogreb database
	val, err := db.Get(key)
	if err != nil {
		return "", err
	}
	return string(val), nil
}

// putDB is to store key & value in badger database
func PutDB(key []byte, value []byte, ifDelete []byte, db *pogreb.DB) error {
	// save the value to pogreb database
	err := db.Put(key, value)
	if err != nil {
		if ifDelete != nil {
			// delete It if there is an error in saving record
			db.Delete(ifDelete)
		}
		return err
	}
	return nil
}

// findBDB is to get the key & value from badger database
func FindBDB(key []byte, bdb *badger.DB) (string, error) {
	var valCopy []byte
	err := bdb.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		} else {
			valCopy, err = item.ValueCopy(nil)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return string(valCopy), nil
}

// CheckBDB is to get the key & value from badger database
func CheckBDB(key []byte, bdb *badger.DB) bool {
	skey := make([]string, 0)
	err := bdb.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		prefix := key
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.Key()
			skey = append(skey, string(k))
			if len(skey) >= 1 {
				break
			}
		}
		return nil
	})
	if err != nil {
		return false
	}
	if len(skey) >= 1 {
		return true
	} else {
		return false
	}
}

// putBDB is to store key & value in badger database
func PutBDB(key []byte, value []byte, ifDelete []byte, bdb *badger.DB) error {
	err := bdb.Update(func(txn *badger.Txn) error {
		err := txn.Set(key, value)
		if err != nil {
			if ifDelete != nil {
				txn.Delete(ifDelete)
			}
		}
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

// parseToken - Parse Token from Authorization Header and get the Token
func ParseToken(authToken string) string {
	if authToken == "" {
		return ""
	}
	bearer := strings.Split(authToken, "Bearer ")
	if len(bearer) != 2 {
		return ""
	}
	token := strings.TrimSpace(bearer[1])
	if len(token) < 1 {
		return ""
	}
	return token
}
