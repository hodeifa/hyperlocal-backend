// Package fcm provides a wrapper for Firebase Admin SDK Cloud Messaging.
package fcm

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

// Config holds the configuration for FCM Client initialization.
type Config struct {
	CredentialsFile string // Path to JSON file (preferred if exists)
	CredentialsJSON string // Fallback: Raw JSON string or Base64
}

// TokenFetcher is an interface to retrieve FCM tokens from the database.
// Concrete implementations will be injected by the service layer (Driver/Customer/OutboxWorker).
type TokenFetcher interface {
	GetCustomerFCMToken(ctx context.Context, customerID string) (string, error)
	GetDriverFCMToken(ctx context.Context, driverID string) (string, error)
}

// fcmSender is an internal interface to facilitate mocking during unit tests.
type fcmSender interface {
	Send(ctx context.Context, message *messaging.Message) (string, error)
}

// Client is the Firebase Admin SDK wrapper.
type Client struct {
	sender  fcmSender
	fetcher TokenFetcher
	logger  *zap.Logger
}

// NewClient initializes the connection to Firebase.
func NewClient(ctx context.Context, cfg Config, fetcher TokenFetcher, logger *zap.Logger) (*Client, error) {
	var opt option.ClientOption

	// Priority 1: File Path
	if cfg.CredentialsFile != "" {
		if _, err := os.Stat(cfg.CredentialsFile); err == nil {
			//nolint:staticcheck // SA1019: deprecated but still standard for simple service account JSON loading in Firebase Admin SDK.
			opt = option.WithCredentialsFile(cfg.CredentialsFile)
		}
	}

	// Priority 2: JSON String / Base64 (Fallback)
	if opt == nil && cfg.CredentialsJSON != "" {
		credJSON := cfg.CredentialsJSON
		// Try decode Base64 if not raw JSON
		if decoded, err := base64.StdEncoding.DecodeString(credJSON); err == nil {
			credJSON = string(decoded)
		}
		//nolint:staticcheck // SA1019: deprecated but still standard for inline JSON loading in Firebase Admin SDK.
		opt = option.WithCredentialsJSON([]byte(credJSON))
	}

	if opt == nil {
		return nil, fmt.Errorf("firebase credentials not found (check FIREBASE_CREDENTIALS_FILE or JSON)")
	}

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to init firebase app: %w", err)
	}

	msgClient, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to init messaging client: %w", err)
	}

	return &Client{
		sender:  msgClient,
		fetcher: fetcher,
		logger:  logger,
	}, nil
}

// Send sends a message directly to an FCM Token.
func (c *Client) Send(ctx context.Context, token, title, body string, data map[string]string) error {
	if token == "" {
		return fmt.Errorf("fcm token cannot be empty")
	}

	msg := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				Title:       title,
				Body:        body,
				ClickAction: "FLUTTER_NOTIFICATION_CLICK",
				Sound:       "default",
			},
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
				},
			},
		},
	}

	responseID, err := c.sender.Send(ctx, msg)
	if err != nil {
		// Truncate token for log security (do not log full PII token)
		tokenPrefix := token
		if len(token) > 10 {
			tokenPrefix = token[:10] + "..."
		}
		c.logger.Error("failed to send FCM",
			zap.String("token_prefix", tokenPrefix),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send FCM: %w", err)
	}

	c.logger.Info("FCM sent successfully", zap.String("response_id", responseID))
	return nil
}

// SendToDriver fetches the driver token from DB then sends the message.
func (c *Client) SendToDriver(ctx context.Context, driverID, title, body, deepLink string) error {
	if c.fetcher == nil {
		return fmt.Errorf("token fetcher not configured")
	}

	token, err := c.fetcher.GetDriverFCMToken(ctx, driverID)
	if err != nil {
		return fmt.Errorf("failed to fetch driver token: %w", err)
	}
	if token == "" {
		c.logger.Debug("driver has no FCM token, skipping", zap.String("driver_id", driverID))
		return nil // Silent skip, not an error
	}

	data := map[string]string{}
	if deepLink != "" {
		data["deep_link"] = deepLink
	}
	return c.Send(ctx, token, title, body, data)
}

// SendToCustomer fetches the customer token from DB then sends the message.
func (c *Client) SendToCustomer(ctx context.Context, customerID, title, body, deepLink string) error {
	if c.fetcher == nil {
		return fmt.Errorf("token fetcher not configured")
	}

	token, err := c.fetcher.GetCustomerFCMToken(ctx, customerID)
	if err != nil {
		return fmt.Errorf("failed to fetch customer token: %w", err)
	}
	if token == "" {
		c.logger.Debug("customer has no FCM token, skipping", zap.String("customer_id", customerID))
		return nil
	}

	data := map[string]string{}
	if deepLink != "" {
		data["deep_link"] = deepLink
	}
	return c.Send(ctx, token, title, body, data)
}
