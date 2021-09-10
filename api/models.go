package api

import (
	"github.com/Bnei-Baruch/archive-my/models"
)

//Responses
type playlistsResponse struct {
	Playlists []*playlistResponse `json:"items"`
	ListResponse
}
type playlistResponse struct {
	*models.Playlist
	ItemsCount    int64                  `json:"count"`
	PlaylistItems []*models.PlaylistItem `json:"playlist_items"`
}
type playlistItemResponse struct {
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

//Requests
type subscribeRequest struct {
	Collections    []string `json:"collections" form:"collections" binding:"omitempty"`
	ContentTypes   []string `json:"types" form:"types" binding:"omitempty"`
	ContentUnitUID string   `json:"content_unit_uid"  binding:"omitempty"`
}

type playListItemRequest struct {
	PlayListIds []string `json:"playlists" binding:"omitempty"`
	UIDsRequest
}

type ListRequest struct {
	PageNumber int    `json:"page_no" form:"page_no" binding:"omitempty,min=1"`
	PageSize   int    `json:"page_size" form:"page_size" binding:"omitempty,min=1"`
	OrderBy    string `json:"order_by" form:"order_by" binding:"omitempty"`
	GroupBy    string `json:"-"`
}

type DeletePlaylistItemRequest struct {
	UIDsRequest
	IDsRequest
	PlaylistId int64 `json:"playlist_id" form:"playlist_id" binding:"omitempty"`
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
