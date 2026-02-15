package firebase

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"cloud.google.com/go/firestore"
	fb "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

// Client is a thin abstraction over the Firebase Admin SDK components
// required by the application. It exposes the Firebase Auth client and
// Firestore client while managing their lifecycle.
type Client struct {
	app *fb.App

	Auth      *auth.Client
	Firestore *firestore.Client

	closeOnce sync.Once
	closeErr  error
}

// Initialize constructs a Firebase Client using the supplied credentials file.
// If credentialsFile is empty, the Firebase Admin SDK falls back to the
// default credential discovery mechanism (e.g. GOOGLE_APPLICATION_CREDENTIALS).
func Initialize(ctx context.Context, credentialsFile string, opts ...option.ClientOption) (*Client, error) {
	if ctx == nil {
		return nil, errors.New("firebase: context must not be nil")
	}

	if credentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(credentialsFile))
	}

	app, err := fb.NewApp(ctx, nil, opts...)
	if err != nil {
		return nil, fmt.Errorf("firebase: create app: %w", err)
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("firebase: initialize auth client: %w", err)
	}

	firestoreClient, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("firebase: initialize firestore client: %w", err)
	}

	return &Client{
		app:       app,
		Auth:      authClient,
		Firestore: firestoreClient,
	}, nil
}

// VerifyIDToken verifies the provided Firebase ID token and returns the decoded token.
func (c *Client) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	if c == nil || c.Auth == nil {
		return nil, errors.New("firebase: auth client is not initialized")
	}
	if idToken == "" {
		return nil, errors.New("firebase: id token is required")
	}
	return c.Auth.VerifyIDToken(ctx, idToken)
}

// GetUser retrieves the Firebase Auth user record for the given UID.
func (c *Client) GetUser(ctx context.Context, uid string) (*auth.UserRecord, error) {
	if c == nil || c.Auth == nil {
		return nil, errors.New("firebase: auth client is not initialized")
	}
	if uid == "" {
		return nil, errors.New("firebase: uid is required")
	}
	return c.Auth.GetUser(ctx, uid)
}

// Close releases any resources held by the Firebase client.
// Currently this closes the Firestore client; additional shutdown logic
// can be added here as needed.
func (c *Client) Close() error {
	if c == nil {
		return nil
	}

	c.closeOnce.Do(func() {
		if c.Firestore != nil {
			c.closeErr = c.Firestore.Close()
		}
	})

	return c.closeErr
}

// App exposes the underlying firebase.App instance for advanced use-cases.
func (c *Client) App() *fb.App {
	if c == nil {
		return nil
	}
	return c.app
}
