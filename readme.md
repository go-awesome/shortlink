## About The Project

<p align="center"><a href="https://github.com/go-awesome/shortlink"><img src="https://repository-images.githubusercontent.com/368965271/37361600-ba7a-11eb-9f5c-966d7a891ce2"></a></p>

Shortlink App in Golang

* Multiple Node based Architecture to create and scale at ease
* Highly performant key-value storage system
* Centralized Storage option when multiple node created - requires tweaking.
* **API auth system not built**. Left for using it for your own use case like `JWT` or `paseto`. Self Implement.

Please see the `architecture` file in the repository on option you can use the app. For some minor tweaking may be required.

### Built With

List of Library and Framework used in building the app:

* [Gofiber](https://gofiber.io)
* [BadgerDB](https://github.com/dgraph-io/badger)
* [PogrebDB](https://github.com/akrylysov/pogreb)
* [hashid](https://github.com/go-awesome/shortlink/blob/main/helper/functions.go#L11)
* [xid](https://github.com/go-awesome/shortlink/blob/main/handler/handler.go#L13)


<!-- GETTING STARTED -->
## Getting Started

Just download and run `go run main.go` and you are ready to go.

### Steps

Common Steps to Launch:

  ```sh
  go mod tidy
  go mod vendor
  go run main.go OR go build -ldflags "-s -w" main.go && ./main
  ```

### Must Changeable Variables in `constant.go`:

```
Production      = 2 // Please set to 1 if in production.
Domain          = "https://lin.ks/"
CookieName      = "lin.ks"
NodeID          = "N1|" // Increase per node by value as "N2|", "N3|"... for multiple node
DBFolder        = "/home/ubuntu/go/src/shortlink/db/"
AddFromToken    = 3 // firt N character to get from token and use it in ShortID
ShortIDToken    = 7 // Further added from 1st N char of AddFromToken+NodeID: total=12
APITokenLength  = 32
```

### Available Routes:

  1. Short URL redirector: `/:short_code_here`
  2. API Routes:
>    - /api/create [Post]
>>     Takes `{"url": "https://github.com"}` with `Authorization: Bearer {token}` from Header
>    - /api/update [Post]
>>     Takes `{"old": "https://github.com", "new": "https://bitbucket.com", "short": "shortcode"}` with `Authorization: Bearer {token}` from Header
>    - /api/delete [Post]
>>     Takes `{ "long": "https://bitbucket.com", "short": "shortcode"}` with `Authorization: Bearer {token}` from Header
>    - /api/fetch [GET]
>>      Takes `Authorization: Bearer {token}` from Header
>    - /api/fetch/:short_code_here [GET]
>>      {short_code_here} in the URL and Takes `Authorization: Bearer {token}` from Header

**Note:** Remember to implement `Auth` system of your own and Replace `APITokenLength` check with your own function.

## Rest API Example:

Please see the `rest.http` file to understand the request type in live details.

## Feature request?

Share your feature request via `issue` tracker.

## Feel like helping out:

- Via Code Contribution (if any / new feature)
- BTC: `1Hp24RtL3o86boapSAD3DtyqF5jdq1rfpM`
- Star the repository and watch out for new updates and features.

<!-- LICENSE -->
## License

Distributed under the Apache License 2.0. See `LICENSE` for more information.
