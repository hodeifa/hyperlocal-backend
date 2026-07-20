// Package fcm provides a wrapper for the Firebase Admin SDK to send push notifications.
package fcm

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

// fcmSender is a thin interface to abstract *messaging.Client.
type fcmSender interface {
	Send(ctx context.Context, message *messaging.Message) (string, error)
}

// TokenFetcher abstracts the database lookup for FCM tokens.
type TokenFetcher interface {
	GetCustomerFCMToken(ctx context.Context, customerID string) (string, error)
	GetDriverFCMToken(ctx context.Context, driverID string) (string, error)
}

// Config holds the Firebase configuration.
type Config struct {
	CredentialsFile string
	CredentialsJSON string
}

// Client is a wrapper for the Firebase Admin SDK.
type Client struct {
	sender  fcmSender
	fetcher TokenFetcher
	logger  *zap.Logger
}

// NewClient initializes the Firebase Admin SDK client.
func NewClient(ctx context.Context, cfg Config, fetcher TokenFetcher, logger *zap.Logger) (*Client, error) {
	// Nil-guard: prevent panic if the caller (or test) forgets to inject the logger.
	if logger == nil {
		logger = zap.NewNop()
	}

	var opts []option.ClientOption
	if cfg.CredentialsFile != "" {
		//nolint:staticcheck // Firebase Admin SDK still requires this function to extract the Project ID.
		// The google.golang.org/api/option package deprecates this, but Firebase has no alternative.
		opts = append(opts, option.WithCredentialsFile(cfg.CredentialsFile))
	} else if cfg.CredentialsJSON != "" {
		//nolint:staticcheck // Same reason as above.
		opts = append(opts, option.WithCredentialsJSON([]byte(cfg.CredentialsJSON)))
	} else {
		return nil, fmt.Errorf("firebase credentials file or JSON must be provided")
	}

	app, err := firebase.NewApp(ctx, nil, opts...)
	if err != nil {
		return nil, fmt.Errorf("error initializing firebase app: %w", err)
	}

	messagingClient, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("error initializing messaging client: %w", err)
	}

	return &Client{
		sender:  messagingClient,
		fetcher: fetcher,
		logger:  logger,
	}, nil
}

// Send sends a push notification. ctx MUST be the first parameter (revive).
func (c *Client) Send(ctx context.Context, token, title, body string, data map[string]string) error {
	if token == "" {
		return fmt.Errorf("target token is empty")
	}

	notification := &messaging.Notification{Title: title, Body: body}
	android := &messaging.AndroidConfig{
		Notification: &messaging.AndroidNotification{
			Title:       title,
			Body:        body,
			ClickAction: "FLUTTER_NOTIFICATION_CLICK",
		},
	}
	apns := &messaging.APNSConfig{
		Payload: &messaging.APNSPayload{
			Aps: &messaging.Aps{
				Alert: &messaging.ApsAlert{Title: title, Body: body},
				Sound: "default",
			},
		},
	}
	if data == nil {
		data = make(map[string]string)
	}

	message := &messaging.Message{
		Token:        token,
		Notification: notification,
		Android:      android,
		APNS:         apns,
		Data:         data,
	}

	response, err := c.sender.Send(ctx, message)
	if err != nil {
		// Safe slice: prevent panic if token < 10 characters (e.g., during unit tests).
		preview := token
		if len(preview) > 10 {
			preview = preview[:10] + "..."
		}
		c.logger.Error("failed to send FCM message", zap.Error(err), zap.String("token_prefix", preview))
		return fmt.Errorf("error sending FCM message: %w", err)
	}

	c.logger.Info("successfully sent FCM message", zap.String("response_id", response))
	return nil
}

// SendToDriver fetches the driver token and sends a notification.
func (c *Client) SendToDriver(ctx context.Context, driverID, title, body, deepLink string) error {
	if c.fetcher == nil {
		return fmt.Errorf("token fetcher not configured")
	}
	token, err := c.fetcher.GetDriverFCMToken(ctx, driverID)
	if err != nil {
		return fmt.Errorf("failed to get driver FCM token: %w", err)
	}
	if token == "" {
		c.logger.Debug("driver has no FCM token, skipping notification", zap.String("driver_id", driverID))
		return nil
	}
	return c.Send(ctx, token, title, body, map[string]string{"deep_link": deepLink})
}

// SendToCustomer fetches the customer token and sends a notification.
func (c *Client) SendToCustomer(ctx context.Context, customerID, title, body, deepLink string) error {
	if c.fetcher == nil {
		return fmt.Errorf("token fetcher not configured")
	}
	token, err := c.fetcher.GetCustomerFCMToken(ctx, customerID)
	if err != nil {
		return fmt.Errorf("failed to get customer FCM token: %w", err)
	}
	if token == "" {
		c.logger.Debug("customer has no FCM token, skipping notification", zap.String("customer_id", customerID))
		return nil
	}
	return c.Send(ctx, token, title, body, map[string]string{"deep_link": deepLink})
}
