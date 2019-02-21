package storage

import (
	"context"
	"os"

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
		DatabaseURL:      os.Getenv("FB_DATABASE_URL"),
		ProjectID:        os.Getenv("FB_PROJECT_ID"),
		ServiceAccountID: os.Getenv("FB_SERVICE_ACC"),
		StorageBucket:    os.Getenv("FB_BUCKET"),
	}
	app, err := firebase.NewApp(ctx, config)
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
	path := os.Getenv("ENV") + "/subscriptionProduct"
	product := &SubscriptionProduct{}
	db, err := initFirebaseDB()
	if err != nil {
		return nil, err
	}

	err = db.Client.NewRef(path).OrderByKey().Get(db.Context, product)
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
// 3. Clean DB each month
