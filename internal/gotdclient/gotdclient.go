// Package gotdclient adds some features to https://github.com/gotd/td library
// like auth client and telegram bot in  and run them in background
package gotdclient

import (
	"context"
	"fmt"

	"github.com/gotd/contrib/bg"
	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram"
)

type Client struct {
	Client *telegram.Client

	stop func() error
}

func New(appID int, appHash string) *Client {
	client := telegram.NewClient(
		appID,
		appHash,
		telegram.Options{
			NoUpdates: true,
		},
	)

	return &Client{
		Client: client,
	}
}

func (c *Client) Connect(ctx context.Context, botToken string) error {
	stop, err := bg.Connect(c.Client)
	if err != nil {
		return fmt.Errorf("failed to connect telegram.Client: %w", err)
	}
	c.stop = stop

	status, err := c.Client.Auth().Status(ctx)
	if err != nil {
		return fmt.Errorf("failed to auth telegram.Client: %w", err)
	}

	if !status.Authorized {
		_, err = c.Client.Auth().Bot(ctx, botToken)
		if err != nil {
			return fmt.Errorf("failed to auth telegram bot with token %q: %w", botToken[:6], err)
		}
	}

	return nil
}

func (c *Client) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	err := c.Client.Invoke(ctx, input, output)
	if err != nil {
		return fmt.Errorf("c.Client.Invoke(ctx, input, output): %w", err)
	}
	return nil
}

func (c *Client) Close() error {
	err := c.stop()
	if err != nil {
		return fmt.Errorf("failed to stop telegram.Client: %w", err)
	}
	return nil
}
