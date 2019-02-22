package storage

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/api/option"

	"firebase.google.com/go/db"

	firebase "firebase.google.com/go"
)

// SubscriptionProduct is the economics invoice ID
type SubscriptionProduct struct {
	Credits  int    `json:"credits,omitempty"`
	StripeID string `json:"id,omitempty"`
	Period   string `json:"period,omitempty"`
}

// DBInstance -
type DBInstance struct {
	Client  *db.Client
	Context context.Context
}

// InitFirebase SE
func initFirebaseDB() (*DBInstance, error) {
	ctx := context.Background()
	config := &firebase.Config{
		DatabaseURL: os.Getenv("FB_DATABASE_URL"),
	}
	opt := option.WithCredentialsJSON(GetSecrets())
	app, err := firebase.NewApp(ctx, config, opt)
	if err != nil {
		panic(err)
	}
	client, err := app.Database(ctx)
	if err != nil {
		return nil, err
	}

	return &DBInstance{
		Client:  client,
		Context: ctx,
	}, nil
}

// GetSubscriptionProducts - this guy
func GetSubscriptionProducts(productNumber string) (*SubscriptionProduct, error) {
	path := os.Getenv("ENV") + "/subscriptionProduct/" + productNumber
	product := &SubscriptionProduct{}
	db, err := initFirebaseDB()
	if err != nil {
		return nil, err
	}
	fmt.Printf("Got the path: %s and db\n", path)

	err = db.Client.NewRef(path).Get(db.Context, product)
	if err != nil {
		return nil, err
	}
	return product, nil
}

// GetCredits specific credit amount pr. product
func (sp *SubscriptionProduct) GetCredits() int {
	return sp.Credits
}

// GetPeriod period ("month"/"year")
func (sp *SubscriptionProduct) GetPeriod() string {
	return sp.Period
}

// 1. Month/year
// 2. Unlimited
