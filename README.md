This is a wrapper around the [Nyckel API](https://www.nyckel.com/docs).

This wrapper only includes the calls that were needed for our first Nyckel project. However, the structure is quite simple and can be reused for any additional
calls that you want.

The file `type.go` defines the types we are using for this API and `nyckel.go` for the implementation.

You'll notice that at the start of `type.go` we have
```
var nyckelClientId = envVarOrPanic("NYCKEL_CLIENT_ID")
var nyckelClientSecret = envVarOrPanic("NYCKEL_CLIENT_SECRET")
```
and this will check that these two environment varibales are set or your program will panic.

The type 
```
type NyckelAPI interface {
...
}
```
is the primary means of talking to the nyckel server once you have an access token (see below).  The API calls in the `NyckelAPI` match the API calls in 
the documentation one to one.

You have to initialize the API with a call to get an access token. You can do this with a function like this:
```
func GetAPIWithAuth() nyckel.NyckelAPI {
	api := &nyckel.NyckelAPIDef{} // temporary, til we get the auth token

	tok := api.AccessTokenOrPanic(time.Now())
	if tok.Expired(time.Now()) {
		log.Fatalf("token is expired, max age was %d", tok.ExpiresSec)
	}
	return nyckel.NewNyckellAPI(tok, true) // now the API is ready
}
```

The returned value from this call is a properly initialized NyckelAPI.  Note that it is user's responsibility to check if the current token is expired and
get a new one if needed.

There is a standard way of reporting errors since the errors returned by Nyckel are encoded in JSON and have some amount of detail. When a golang "error" 
is returned by an API call you can use a type assertion to convert it to a `NyckelCallResult` which has some extra information about the call such as the
HTTP error code and the error message as json.
