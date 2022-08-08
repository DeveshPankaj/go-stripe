package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/paymentintent"
	"github.com/stripe/stripe-go/refund"
	"io"
	"log"
	"net/http"
	"os"
)



var STRIPE_KEY string
var HOST = ":4000"



type CreateIntentPayload struct {
	Amount      int64 `json:"amount"`
	Currency    string `json:"currency"`
	Description string `json:"description"`
}

func createIntent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	stripe.Key = STRIPE_KEY

	var payload CreateIntentPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	params := &stripe.PaymentIntentParams{
		Amount: stripe.Int64(payload.Amount),
		Currency: stripe.String(payload.Currency),
		Description: stripe.String(payload.Description),
		CaptureMethod: stripe.String(string(stripe.PaymentIntentCaptureMethodManual)),
		PaymentMethodTypes: []*string {
			stripe.String("card"),
		},
	}

	pi, _ := paymentintent.New(params)

	json.NewEncoder(w).Encode(struct {
		ClientSecret string `json:"clientSecret"`
		PaymentIntent string `json:"payment_intent"`
	}{
		ClientSecret: pi.ClientSecret,
		PaymentIntent: pi.ID,
	})
}


func captureIntent(w http.ResponseWriter, r *http.Request) {
	stripe.Key = STRIPE_KEY
	w.Header().Set("Content-type", "application/json")
	params := mux.Vars(r)
	id := params["pi"]
	// To create a requires_capture PaymentIntent, see our guide at: https://stripe.com/docs/payments/capture-later
	pi, err := paymentintent.Capture(
		id,
		nil,
	)

	if err != nil {
		json.NewEncoder(w).Encode(err)
		return
	}

	json.NewEncoder(w).Encode(pi)
}


func confirmPayment(w http.ResponseWriter, req * http.Request)  {
	stripe.Key = STRIPE_KEY

	query := mux.Vars(req)
	id := query["pi"]

	params := &stripe.PaymentIntentConfirmParams{
		PaymentMethod: stripe.String("pm_card_visa"),
	}
	pi, _ := paymentintent.Confirm(
		id,
		params,
	)

	json.NewEncoder(w).Encode(pi)
}

func getIntent(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	stripe.Key = STRIPE_KEY

	pi, _ := paymentintent.Get(
		id,
		nil,
	)

	json.NewEncoder(w).Encode(pi)
}
//
//func handleWebhook(w http.ResponseWriter, req *http.Request) {
//	const MaxBodyBytes = int64(65536)
//	req.Body = http.MaxBytesReader(w, req.Body, MaxBodyBytes)
//	payload, err := ioutil.ReadAll(req.Body)
//	if err != nil {
//		fmt.Fprintf(os.Stderr, "Error reading request body: %v\n", err)
//		w.WriteHeader(http.StatusServiceUnavailable)
//		return
//	}
//
//	// This is your Stripe CLI webhook secret for testing your endpoint locally.
//	endpointSecret := "whsec_<>"
//	// Pass the request body and Stripe-Signature header to ConstructEvent, along
//	// with the webhook signing key.
//	event, err := webhook.ConstructEvent(payload, req.Header.Get("Stripe-Signature"),
//		endpointSecret)
//
//	if err != nil {
//		fmt.Fprintf(os.Stderr, "Error verifying webhook signature: %v\n", err)
//		w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
//		return
//	}
//
//	// Unmarshal the event data into an appropriate struct depending on its Type
//	switch event.Type {
//	case "payment_intent.succeeded":
//		// Then define and call a function to handle the event payment_intent.succeeded
//	// ... handle other event types
//	default:
//		fmt.Fprintf(os.Stderr, "Unhandled event type: %s\n", event.Type)
//	}
//
//	w.WriteHeader(http.StatusOK)
//}

func getIntents(w http.ResponseWriter, req *http.Request) {
	stripe.Key = STRIPE_KEY

	params := &stripe.PaymentIntentListParams{}
	params.Filters.AddFilter("limit", "", "3")
	i := paymentintent.List(params)

	var ls []*stripe.PaymentIntent
	for i.Next() {
		pi := i.PaymentIntent()
		ls = append(ls, pi)
	}

	writeJSON(w, ls)

}

func writeJSON(w http.ResponseWriter, v interface{}) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("json.NewEncoder.Encode: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := io.Copy(w, &buf); err != nil {
		log.Printf("io.Copy: %v", err)
		return
	}
}

func createRefund(w http.ResponseWriter, req *http.Request) {
	stripe.Key = STRIPE_KEY

	args := mux.Vars(req)
	id := args["pi"]

	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(id),
		//Amount: stripe.Int64(1000),
	}
	result, err := refund.New(params)

	if err != nil {
		writeJSON(w, err)
		return
	}

	json.NewEncoder(w).Encode(struct {
		status string `json:"status"`
	}{
		status: string(result.Status),
	})
}


func main() {
	godotenv.Load()
	r := mux.NewRouter()

	STRIPE_KEY = os.Getenv("STRIPE_KEY")

	if len(STRIPE_KEY) == 0 {
		fmt.Errorf("STRIPE_KEY not fount in environment variables!")
		return
	}

	r.HandleFunc("/api/v1/create_intent", createIntent).Methods("POST")
	r.HandleFunc("/api/v1/get_intent/{id}", getIntent).Methods("GET")
	r.HandleFunc("/api/v1/get_intents", getIntents).Methods("GET")
	r.HandleFunc("/api/v1/confirm/{pi}", confirmPayment).Methods("GET")
	r.HandleFunc("/api/v1/capture_intent/{pi}", captureIntent).Methods("POST")
	r.HandleFunc("/api/v1/create_refund/{pi}", createRefund).Methods("POST")

	//r.HandleFunc("/api/webhook", handleWebhook).Methods("GET", "POST", "PUT")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))

	fmt.Printf("Server is running at %s\n", HOST)
	log.Fatal(http.ListenAndServe(HOST, r))
}
