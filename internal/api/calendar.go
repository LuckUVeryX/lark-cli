package api

import (
	"fmt"
)

// GetPrimaryCalendar retrieves the user's primary calendar
func (c *Client) GetPrimaryCalendar() (*Calendar, error) {
	var resp CalendarResponse

	if err := c.Post("/calendar/v4/calendars/primary", nil, &resp); err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("API error (code %d): %s", resp.Code, resp.Msg)
	}

	// Primary calendar returns an array of calendars
	if len(resp.Data.Calendars) > 0 && resp.Data.Calendars[0].Calendar != nil {
		return resp.Data.Calendars[0].Calendar, nil
	}

	// Fallback for single calendar response
	if resp.Data.Calendar != nil {
		return resp.Data.Calendar, nil
	}

	return nil, fmt.Errorf("no calendar data in response")
}

// GetCalendar retrieves a specific calendar by ID
func (c *Client) GetCalendar(calendarID string) (*Calendar, error) {
	var resp CalendarResponse

	path := fmt.Sprintf("/calendar/v4/calendars/%s", calendarID)
	if err := c.Get(path, &resp); err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("API error (code %d): %s", resp.Code, resp.Msg)
	}

	return resp.Data.Calendar, nil
}

// ListCalendars retrieves all calendars for the user
func (c *Client) ListCalendars() ([]Calendar, error) {
	var allCalendars []Calendar
	var pageToken string

	for {
		var resp CalendarListResponse
		path := "/calendar/v4/calendars"
		if pageToken != "" {
			path += "?page_token=" + pageToken
		}

		if err := c.Get(path, &resp); err != nil {
			return nil, err
		}

		if resp.Code != 0 {
			return nil, fmt.Errorf("API error (code %d): %s", resp.Code, resp.Msg)
		}

		allCalendars = append(allCalendars, resp.Data.Calendars...)

		if !resp.Data.HasMore {
			break
		}
		pageToken = resp.Data.PageToken
	}

	return allCalendars, nil
}
