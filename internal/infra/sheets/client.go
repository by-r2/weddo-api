package sheets

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	gsheets "google.golang.org/api/sheets/v4"
)

type Client struct {
	service       *gsheets.Service
	spreadsheetID string
}

func NewClientFromTokenSource(ctx context.Context, spreadsheetID string, ts oauth2.TokenSource) (*Client, error) {
	if spreadsheetID == "" {
		return nil, fmt.Errorf("sheets.NewClientFromTokenSource: spreadsheet id vazio")
	}
	svc, err := gsheets.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("sheets.NewClientFromTokenSource: create service: %w", err)
	}
	return &Client{service: svc, spreadsheetID: spreadsheetID}, nil
}

func NewRawServiceFromTokenSource(ctx context.Context, ts oauth2.TokenSource) (*gsheets.Service, error) {
	svc, err := gsheets.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("sheets.NewRawServiceFromTokenSource: %w", err)
	}
	return svc, nil
}

func CreateSpreadsheet(ctx context.Context, service *gsheets.Service, title string) (id, url string, err error) {
	resp, err := service.Spreadsheets.Create(&gsheets.Spreadsheet{
		Properties: &gsheets.SpreadsheetProperties{
			Title: title,
		},
	}).Context(ctx).Do()
	if err != nil {
		return "", "", fmt.Errorf("sheets.CreateSpreadsheet: %w", err)
	}
	return resp.SpreadsheetId, resp.SpreadsheetUrl, nil
}

func (c *Client) WriteTab(ctx context.Context, tabName string, values [][]any) error {
	if err := c.ensureTab(ctx, tabName); err != nil {
		return err
	}

	if len(values) == 0 {
		values = [][]any{{}}
	}

	valueRange := &gsheets.ValueRange{Values: values}
	rangeAll := fmt.Sprintf("%s!A:Z", tabName)

	_, err := c.service.Spreadsheets.Values.Clear(c.spreadsheetID, rangeAll, &gsheets.ClearValuesRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("sheets.WriteTab: clear %s: %w", tabName, err)
	}

	_, err = c.service.Spreadsheets.Values.Update(c.spreadsheetID, rangeAll, valueRange).
		ValueInputOption("RAW").
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("sheets.WriteTab: update %s: %w", tabName, err)
	}

	return nil
}

func (c *Client) ReadTab(ctx context.Context, tabName string) ([][]string, error) {
	if err := c.ensureTab(ctx, tabName); err != nil {
		return nil, err
	}

	rangeAll := fmt.Sprintf("%s!A:Z", tabName)
	resp, err := c.service.Spreadsheets.Values.Get(c.spreadsheetID, rangeAll).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("sheets.ReadTab: get %s: %w", tabName, err)
	}

	rows := make([][]string, 0, len(resp.Values))
	for _, row := range resp.Values {
		out := make([]string, 0, len(row))
		for _, cell := range row {
			out = append(out, fmt.Sprintf("%v", cell))
		}
		rows = append(rows, out)
	}

	return rows, nil
}

func (c *Client) ensureTab(ctx context.Context, tabName string) error {
	metadata, err := c.service.Spreadsheets.Get(c.spreadsheetID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("sheets.ensureTab: get spreadsheet: %w", err)
	}

	for _, s := range metadata.Sheets {
		if s.Properties != nil && s.Properties.Title == tabName {
			return nil
		}
	}

	req := &gsheets.BatchUpdateSpreadsheetRequest{
		Requests: []*gsheets.Request{
			{
				AddSheet: &gsheets.AddSheetRequest{
					Properties: &gsheets.SheetProperties{
						Title: tabName,
					},
				},
			},
		},
	}

	_, err = c.service.Spreadsheets.BatchUpdate(c.spreadsheetID, req).Context(ctx).Do()
	if err != nil {
		var gerr *googleapi.Error
		if errors.As(err, &gerr) && gerr.Code == http.StatusConflict {
			return nil
		}
		return fmt.Errorf("sheets.ensureTab: add sheet %s: %w", tabName, err)
	}

	return nil
}
