package main

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/go-awesome/shortlink/handler"
	"github.com/go-awesome/shortlink/helper"

	"github.com/akrylysov/pogreb"
	badger "github.com/dgraph-io/badger/v3"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	db, err := pogreb.Open(helper.DBFolder+"/pogreb/url.db", nil)
	if err != nil {
		log.Fatal(helper.ErrorPrint(err.Error(), helper.DB101))
		return
	}
	defer db.Close()
	opts := badger.DefaultOptions(helper.DBFolder + "/badger")
	opts.NumVersionsToKeep = 1
	opts.ReadOnly = false

	if helper.BypassLockGuard {
		// When uisng centralized DB server, it will be read from many and write to 1 node.
		opts.BypassLockGuard = true
	}

	// Open badger Database
	bdb, err := badger.Open(opts)
	if err != nil {
		log.Fatal(helper.ErrorPrint(err.Error(), helper.DB102))
	}
	defer bdb.Close()

	// Close the database on interruptor
	interruptor := make(chan os.Signal, 1)
	signal.Notify(interruptor, os.Interrupt)
	go func() {
		for range interruptor {
			db.Close()
			bdb.Close()
			os.Exit(1)
		}
	}()

	var n int = 1

	// Start the Gofiber APP
	app := fiber.New(fiber.Config{
		ReadTimeout:           10 * time.Minute,
		WriteTimeout:          5 * time.Minute,
		Prefork:               false,
		CaseSensitive:         false,
		StrictRouting:         true,
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": "true", "Message": helper.ErrorPrint(err.Error(), helper.CO101)})
		},
	})
	app.Use(recover.New(), favicon.New(favicon.Config{
		File: "./favicon.ico",
	}))

	/**
	 * fetch a short ID and redirect
	 * @param `shortURL` from URL
	 * @action also store analytics of the URL
	 * @return redirect
	**/

	api := app.Group("/api")

	// IndexHandler
	api.Get("/", handler.IndexHandler)

	/**
	 * Store a new Long URL in storage.
	 * @param `url` from JSON
	 * @action generate unique short ID, increment n, store short id.
	 * @return json response with error and successful message
	**/

	api.Post("/create", handler.CreateHandler(n, bdb, db))

	/**
	 * Update a Long URL in storage.
	 * @param `url` from JSON & Authorization token from Header
	 * @action update database for long URL
	 * @return json response with error and successful message
	 */

	api.Post("/update", handler.UpdateHandler(bdb, db))

	/**
	 * Fetch a list of Long URL in storage for the API Holder.
	 * @param `Authorization Token` from Header
	 * @action fetch entry list from database
	 * @return json response with error and successful message with list
	 */

	api.Get("/fetch", handler.FetchAllHandler(bdb, db))

	/**
	 * Fetch detail analytics in storage by the API holder
	 * @param `ShortID` from URL
	 * @action fetch all entry from database
	 * @return json response with error and successful message
	**/

	api.Get("/fetch/:code", handler.FetchSingleHandler(bdb, db))

	/**
	 * Delete a Long URL in storage by the API holder
	 * @param `helper.DeleteURL struct` from JSON
	 * @action delete entry from database
	 * @return json response with error and successful message
	**/
	api.Post("/delete", handler.DeleteHandler(bdb, db))

	app.Get("/:code", handler.RedirectToMeWebsite(db, bdb))

	// run the server...
	app.Listen(":8080")
}
