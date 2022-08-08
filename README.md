
# build app 
go build .


# Config
Set environment variables in `.env`
or export from shell
```shell
$ export STRIPE_KEY=<stripe account key>
```

# run app
./go-stripe


# APIs
## Create intent for payment
    POST `/api/v1/create_intent`


## Capture the created intent 
    POST `/api/v1/capture_intent/:payment_intent`
  

## Create a refund for the created intent
    POST `/api/v1/create_refund/:payment_intent`
  

## Get a List of all intents
    GET `/api/v1/get_intents`


## Payment UI
    /checkout.html?clientSecret=<[`clientSecret`](#APIs)>
  
