package handler

import (
	"encoding/json"
	"shortlink/helper"
	"strconv"
	"strings"
	"time"

	"github.com/akrylysov/pogreb"
	"github.com/dgraph-io/badger/v3"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/xid"
)

func IndexHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"node": helper.NodeID, "message": "Short URL Service Provider"})
}

func CreateHandler(n int, bdb *badger.DB, db *pogreb.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		post := new(helper.CreateURL)
		if err := c.BodyParser(&post); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "true", "message": helper.ErrorPrint(err.Error(), helper.CO101)})
		}
		// should do better URL validation here
		if !helper.ValidateURL(post.URL) {
			return c.Status(400).JSON(fiber.Map{"error": "true", "message": "Invalid URL | Minimum URL Length should be 5"})
		}
		// Get the API Token from Authorization Bearer Header
		token := helper.ParseToken(c.Get("Authorization"))

		// Basic Checking without Real time API Token checking
		if len(token) != helper.APITokenLength {
			return c.Status(400).JSON(fiber.Map{"error": "true", "message": helper.ErrorPrint(helper.ID103, helper.ID103)})
		}
		msgTime := time.Now().Format("2006-01-02-15:04:05")
		// Create MD5 Hash of URL
		md5URL := helper.CreateMD5Hash(post.URL)
		// Check if user already has the URL
		val, err := helper.FindBDB([]byte(token+"-"+md5URL), bdb)
		if err == nil {
			return c.JSON(fiber.Map{"error": "false", "message": val})
		}
		// Checking increment of N variable for new ShortID Token
		if n == 1 {
			num, err := helper.FindBDB([]byte("lastID"), bdb)
			if err != nil && err.Error() == "Key not found" {
				// do nothing as it is first entry
			} else if err != nil && err.Error() != "Key not found" {
				return c.Status(400).JSON(fiber.Map{"error": "true", "Message": helper.ErrorPrint(err.Error(), helper.ID101)})
			} else {
				byteToInt, err := strconv.Atoi(num)
				if err != nil {
					return c.Status(400).JSON(fiber.Map{"error": "true", "Message": helper.ErrorPrint(err.Error(), helper.ID102)})
				}
				if byteToInt > 1 {
					n = byteToInt + 1
				}
			}
		}

		// Generate Short Token ID for given index n
		shortID, err := helper.GenerateToken(token, n)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "true", "message": helper.ErrorPrint(err.Error(), helper.ID104)})
		}

		// json for storage
		jstore, err := json.Marshal(fiber.Map{"createdAt": msgTime, "shortURL": helper.Domain + shortID, "longURL": post.URL})
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "true", "message": helper.ErrorPrint(err.Error(), helper.ID104)})
		}
		/**
		* Now Create 2 KEY "API+md of URL" and "shortID"
		* One key will be used to populate list of all URL done by API Holder
		* Second key will be for redirect
		 */
		// Stored in Pogreb Database
		err = helper.PutDB([]byte(shortID), []byte(post.URL), nil, db)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "true", "message": helper.ErrorPrint(err.Error(), helper.ID104)})
		}

		// Stored in Badger Database
		err = helper.PutBDB([]byte(token+"-"+md5URL), jstore, nil, bdb)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "true", "message": helper.ErrorPrint(err.Error(), helper.ID104)})
		}
		// Increment the last ID and update the database
		err = helper.PutBDB([]byte("lastID"), []byte(strconv.Itoa(n)), []byte(token+"-"+md5URL), bdb)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "true", "message": helper.ErrorPrint(err.Error(), helper.ID104)})
		}
		n++
		return c.JSON(fiber.Map{"error": "false", "message": string(jstore)})
	}
}

func FetchAllHandler(bdb *badger.DB, db *pogreb.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := helper.ParseToken(c.Get("Authorization"))
		// Basic Checking without Real time API Token checking
		if len(token) != helper.APITokenLength {
			return c.Status(400).JSON(fiber.Map{"error": "true", "message": helper.ID103})
		}
		skey := make([]string, 0)
		bdb.View(func(txn *badger.Txn) error {
			it := txn.NewIterator(badger.DefaultIteratorOptions)
			defer it.Close()
			prefix := []byte(token + "-")
			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				item := it.Item()
				// k := item.Key()
				err := item.Value(func(v []byte) error {
					skey = append(skey, string(v))
					return nil
				})
				if err != nil {
					return err
				}
			}
			return nil
		})
		jstore, _ := json.Marshal(skey)
		return c.JSON(fiber.Map{"error": "false", "message": string(jstore)})
	}
}

func UpdateHandler(bdb *badger.DB, db *pogreb.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		post := new(helper.UpdateURL)
		if err := c.BodyParser(&post); err != nil {
			return c.Status(400).Send([]byte(err.Error()))
		}
		if !helper.ValidateURL(post.NURL) {
			return c.JSON(fiber.Map{"error": "true", "message": "Invalid New URL"})
		}
		// Get the API Token from Authorization Bearer Header
		token := helper.ParseToken(c.Get("Authorization"))

		// Basic Checking without Real time API Token checking
		if len(token) != helper.APITokenLength {
			return c.Status(400).JSON(fiber.Map{"error": "true", "message": helper.ErrorPrint(helper.ID103, helper.ID103)})
		}
		md5URL := helper.CreateMD5Hash(post.OURL)
		// Check if user already has the URL
		oldVal, err := helper.FindBDB([]byte(token+"-"+md5URL), bdb)
		if err != nil {
			return c.JSON(fiber.Map{"error": "true", "message": helper.ErrorPrint(err.Error(), helper.ID108)})
		}

		val, err := helper.FindDB([]byte(post.ShortID), db)
		if err != nil {
			return c.JSON(fiber.Map{"error": "true", "message": helper.ErrorPrint(err.Error(), helper.ID105)})
		}
		if val != post.OURL {
			return c.Status(400).JSON(fiber.Map{"error": "true", "message": helper.ErrorPrint(helper.ID103, helper.ID108)})
		}

		Newmd5URL := helper.CreateMD5Hash(post.NURL)
		// Check if user already has the URL
		_, err = helper.FindBDB([]byte(token+"-"+Newmd5URL), bdb)
		if err == nil {
			return c.JSON(fiber.Map{"error": "true", "message": helper.ID109})
		}
		// Update OLD Database
		err = helper.PutBDB([]byte(token+"-"+Newmd5URL), []byte(oldVal), []byte(token+"-"+md5URL), bdb)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "true", "message": helper.ErrorPrint(err.Error(), helper.ID104)})
		}
		// Update New Redirect
		err = helper.PutDB([]byte(post.ShortID), []byte(post.NURL), nil, db)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "true", "message": helper.ErrorPrint(err.Error(), helper.ID104)})
		}

		return c.JSON(fiber.Map{"error": "false", "message": "Successfully Updated"})
	}
}

func FetchSingleHandler(bdb *badger.DB, db *pogreb.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		code := c.Params("code")
		code = strings.Replace(code, "%7C", "|", -1)
		if len(code) > helper.ShortIDToken {
			token := helper.ParseToken(c.Get("Authorization"))
			// Basic Checking without Real time API Token checking
			if len(token) != helper.APITokenLength {
				return c.Status(400).JSON(fiber.Map{"error": "true", "message": helper.ID103})
			}
			val, err := helper.FindDB([]byte(code), db)
			if err != nil {
				return c.JSON(fiber.Map{"error": "true", "message": helper.ErrorPrint(err.Error(), helper.ID105)})
			}
			traffic := make(map[string]int)
			table := make(map[string]string)
			bdb.View(func(txn *badger.Txn) error {
				opts := badger.DefaultIteratorOptions
				opts.PrefetchValues = false
				it := txn.NewIterator(opts)
				defer it.Close()
				prefix := []byte(code + "-")
				for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
					item := it.Item()
					k := item.Key()
					split := strings.Split(string(k), "-|-")
					if len(split) == 3 {
						if _, present := traffic["total"]; !present {
							traffic["total"] = 1
						} else {
							traffic["total"] += 1
						}
						timer := strings.Split(split[2], "-")
						if _, present := table[split[1]]; !present {
							table[split[1]] = "done"
							traffic["unique"] += 1
							if _, present := traffic["unique-"+timer[0]+"-"+timer[1]+"-"+timer[2]]; !present {
								traffic["unique-"+timer[0]+"-"+timer[1]+"-"+timer[2]] = 1
							} else {
								traffic["unique-"+timer[0]+"-"+timer[1]+"-"+timer[2]] += 1
							}
						}
						if _, present := traffic[timer[0]+"-"+timer[1]+"-"+timer[2]]; !present {
							traffic[timer[0]+"-"+timer[1]+"-"+timer[2]] = 1
						} else {
							traffic[timer[0]+"-"+timer[1]+"-"+timer[2]] += 1
						}
					}
				}
				return nil
			})
			return c.JSON(fiber.Map{"error": "false", "message": traffic, "longurl": val})
		}
		return c.JSON(fiber.Map{"error": "true", "message": "Invalid Short Code Length"})
	}
}

func DeleteHandler(bdb *badger.DB, db *pogreb.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		post := new(helper.DeleteURL)
		if err := c.BodyParser(&post); err != nil {
			return c.Status(400).Send([]byte(err.Error()))
		}
		if !helper.ValidateURL(post.URL) {
			return c.JSON(fiber.Map{"error": "true", "message": "Invalid URL"})
		}
		// Get the API Token from Authorization Bearer Header
		token := helper.ParseToken(c.Get("Authorization"))

		// Basic Checking without Real time API Token checking
		if len(token) != helper.APITokenLength {
			return c.Status(400).JSON(fiber.Map{"error": "true", "message": helper.ErrorPrint(helper.ID103, helper.ID103)})
		}
		md5URL := helper.CreateMD5Hash(post.URL)
		// Check if user already has the URL
		_, err := helper.FindBDB([]byte(token+"-"+md5URL), bdb)
		if err != nil {
			return c.JSON(fiber.Map{"error": "true", "message": helper.ErrorPrint(err.Error(), helper.ID107)})
		}
		// delete short URL
		err = db.Delete([]byte(post.ShortID))
		if err != nil {
			return c.JSON(fiber.Map{"error": "true", "message": helper.ErrorPrint(err.Error(), helper.ID107)})
		}
		// delete "token-md5url" + "analytics" data
		err = bdb.Update(func(txn *badger.Txn) error {
			err = txn.Delete([]byte(token + "-" + md5URL))
			it := txn.NewIterator(badger.DefaultIteratorOptions)
			defer it.Close()
			prefix := []byte(post.ShortID + "-")
			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				item := it.Item()
				k := item.Key()
				txn.Delete(k)
			}
			return err
		})
		if err != nil {
			return c.JSON(fiber.Map{"error": "true", "message": helper.ErrorPrint(err.Error(), helper.ID107)})
		}
		return c.JSON(fiber.Map{"error": "false", "message": "Successfully deleted"})
	}
}

func RedirectToMeWebsite(db *pogreb.DB, bdb *badger.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		msgTime := time.Now().Format("2006-01-02-15:04:05")
		code := c.Params("code")
		if len(code) > helper.ShortIDToken {
			code = strings.Replace(code, "%7C", "|", -1)
			val, err := helper.FindDB([]byte(code), db)
			if err != nil {
				return c.JSON(fiber.Map{"error": "true", "message": helper.ErrorPrint(err.Error(), helper.ID105)})
			}
			if !helper.ValidateURL(val) {
				return c.JSON(fiber.Map{"error": "true", "message": helper.ID106})
			}
			// Do Cookie Way + IP Way
			getCookie := c.Cookies(helper.CookieName, "no")
			if len(getCookie) > 2 {
				checkIfExist := helper.CheckBDB([]byte(code+"-|-"+getCookie+"-|-"), bdb)
				if !checkIfExist {
					guid := xid.New()
					getCookie = guid.String()
					c.Cookie(&fiber.Cookie{
						Name:     helper.CookieName,
						Value:    getCookie,
						Expires:  time.Now().Add(365 * 24 * time.Hour),
						HTTPOnly: true,
						SameSite: "strict",
					})
				}
				helper.PutBDB([]byte(code+"-|-"+getCookie+"-|-"+msgTime), nil, nil, bdb)
			} else {
				guid := xid.New()
				cookieToken := guid.String()
				c.Cookie(&fiber.Cookie{
					Name:     helper.CookieName,
					Value:    cookieToken,
					Expires:  time.Now().Add(365 * 24 * time.Hour),
					HTTPOnly: true,
					SameSite: "strict",
				})
				// when we will be spliting the string with "|", we will have node,code,ip,date
				helper.PutBDB([]byte(code+"-|-"+cookieToken+"-|-"+msgTime), nil, nil, bdb)
			}
			return c.Redirect(val)
		}
		return c.JSON(fiber.Map{"error": "true", "message": helper.ID106})
	}
}
