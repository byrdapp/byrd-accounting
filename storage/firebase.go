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
	Credits  int    `json:"credits"`
	StripeID string `json:"id"`
	Period   string `json:"period"`
}

// DBInstance -
type DBInstance struct {
	Client  *db.Client
	Context context.Context
}

// InitFirebaseDB SE
func InitFirebaseDB() (*DBInstance, error) {
	ctx := context.Background()
	config := &firebase.Config{
		DatabaseURL: os.Getenv("FB_DATABASE_URL"),
	}
	jsonPath := "fb-" + os.Getenv("ENV") + ".json"
	opt := option.WithCredentialsJSON(GetAWSSecrets(jsonPath))
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
func GetSubscriptionProducts(db *DBInstance, productNumber string) (*SubscriptionProduct, error) {
	path := os.Getenv("ENV") + "/subscriptionProducts/" + productNumber
	product := SubscriptionProduct{}
	fmt.Printf("Path: %s\n", path)
	ref := db.Client.NewRef(path)
	if err := ref.Get(db.Context, &product); err != nil {
		return nil, err
	}
	return &product, nil
}
