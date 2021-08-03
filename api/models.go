package api

import "archive-my/models"

type playlistsResponse struct {
	Playlist []*models.Playlist `json:"playlist"`
	ListResponse
}

type likesResponse struct {
	Likes []*models.Like `json:"likes"`
	ListResponse
}

type subscriptionsResponse struct {
	Subscriptions []*models.Subscription `json:"subscriptions"`
	ListResponse
}
type historyResponse struct {
	History []*models.History `json:"history"`
	ListResponse
}

type subscribeRequest struct {
	Collections  []string `json:"collections" form:"collections" binding:"omitempty"`
	ContentTypes []int64  `json:"types" form:"types" binding:"omitempty"`
}

type ListRequest struct {
	PageNumber int    `json:"page_no" form:"page_no" binding:"omitempty,min=1"`
	PageSize   int    `json:"page_size" form:"page_size" binding:"omitempty,min=1"`
	StartIndex int    `json:"start_index" form:"start_index" binding:"omitempty,min=1"`
	StopIndex  int    `json:"stop_index" form:"stop_index" binding:"omitempty,min=1"`
	OrderBy    string `json:"order_by" form:"order_by" binding:"omitempty"`
	GroupBy    string `json:"-"`
}

type ListResponse struct {
	Total int64 `json:"total"`
}
