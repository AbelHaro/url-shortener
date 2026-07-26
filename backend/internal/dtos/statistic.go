package dtos

import "time"

// ResolveShortURLRequest contains optional client context for a short-link resolution.
// @name ResolveShortURLRequest
type ResolveShortURLRequest struct {
	Referrer string `json:"referrer"`
}

// ResolveShortURLResponse contains the destination for a shortened URL.
// @name ResolveShortURLResponse
type ResolveShortURLResponse struct {
	OriginalURL string `json:"original_url" binding:"required"`
}

// DailyClickResponse contains the click count for one UTC calendar day.
// @name DailyClickResponse
type DailyClickResponse struct {
	Date   string `json:"date" binding:"required"`
	Clicks int64  `json:"clicks"`
}

// ReferrerCountResponse contains an aggregated referrer click count.
// @name ReferrerCountResponse
type ReferrerCountResponse struct {
	Referrer string `json:"referrer" binding:"required"`
	Clicks   int64  `json:"clicks"`
}

// RecentClickResponse contains a recent click event.
// @name RecentClickResponse
type RecentClickResponse struct {
	ClickedAt time.Time `json:"clicked_at" binding:"required"`
	Referrer  string    `json:"referrer" binding:"required"`
}

// URLStatisticsResponse contains the analytics dashboard for one shortened URL.
// @name URLStatisticsResponse
type URLStatisticsResponse struct {
	URL           URLResponse             `json:"url" binding:"required"`
	TotalClicks   int64                   `json:"total_clicks"`
	LastClickedAt *time.Time              `json:"last_clicked_at"`
	ClicksByDay   []DailyClickResponse    `json:"clicks_by_day" binding:"required"`
	TopReferrers  []ReferrerCountResponse `json:"top_referrers" binding:"required"`
	RecentClicks  []RecentClickResponse   `json:"recent_clicks" binding:"required"`
}
