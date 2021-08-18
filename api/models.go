package api

import "archive-my/models"

type playlistsResponse struct {
	Playlist []*playlistResponse `json:"items"`
	ListResponse
}

type playlistResponse struct {
	*models.Playlist
	PlaylistItems []*models.PlaylistItem `json:"items"`
}

type likesResponse struct {
	Likes []*models.Like `json:"items"`
	ListResponse
}

type subscriptionsResponse struct {
	Subscriptions []*models.Subscription `json:"items"`
	ListResponse
}
type historyResponse struct {
	History []*models.History `json:"items"`
	ListResponse
}

type subscribeRequest struct {
	Collections    []string `json:"collections" form:"collections" binding:"omitempty"`
	ContentTypes   []string `json:"types" form:"types" binding:"omitempty"`
	ContentUnitUID string   `json:"content_unit_uid"  binding:"omitempty"`
}

type ListRequest struct {
	PageNumber int    `json:"page_no" form:"page_no" binding:"omitempty,min=1"`
	PageSize   int    `json:"page_size" form:"page_size" binding:"omitempty,min=1"`
	OrderBy    string `json:"order_by" form:"order_by" binding:"omitempty"`
	GroupBy    string `json:"-"`
}

type UIDsRequest struct {
	UIDs []string `json:"uids" form:"uids" binding:"omitempty"`
}

type IDsRequest struct {
	IDs []int64 `json:"ids" form:"ids" binding:"omitempty"`
}

type ListResponse struct {
	Total int64 `json:"total"`
}
