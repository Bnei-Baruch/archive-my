package api

import (
	"time"

	"github.com/volatiletech/null/v8"
)

type ListRequest struct {
	PageNumber int    `json:"page_no" form:"page_no" binding:"omitempty,min=1"`
	PageSize   int    `json:"page_size" form:"page_size" binding:"omitempty,min=1"`
	OrderBy    string `json:"order_by" form:"order_by" binding:"omitempty"`
	GroupBy    string `json:"-"`
}

type ListResponse struct {
	Total int64 `json:"total"`
}

type IDsFilter struct {
	IDs []int64 `json:"ids" form:"ids" binding:"omitempty"`
}

type UIDsFilter struct {
	UIDs []string `json:"uids" form:"uids" binding:"omitempty,dive,len=8"`
}

type GetPlaylistsRequest struct {
	ListRequest
}

type PlaylistsResponse struct {
	ListResponse
	Items []*Playlist `json:"items"`
}

func NewPlaylistsResponse(total int64, numItems int) *PlaylistsResponse {
	return &PlaylistsResponse{
		ListResponse: ListResponse{Total: total},
		Items:        make([]*Playlist, numItems)}
}

type PlaylistRequest struct {
	Name       string                 `json:"name" binding:"omitempty,max=256"`
	Public     bool                   `json:"public"`
	Properties map[string]interface{} `json:"properties" binding:"omitempty"`
}

type PlaylistItemAddInfo struct {
	Position       int    `json:"position" binding:"required"`
	ContentUnitUID string `json:"content_unit_uid" binding:"required,len=8"`
}

type AddPlaylistItemsRequest struct {
	Items []PlaylistItemAddInfo `json:"items" binding:"required,dive"`
}

type PlaylistItemUpdateInfo struct {
	PlaylistItemAddInfo
	ID int64 `json:"id" binding:"omitempty"`
}

type UpdatePlaylistItemsRequest struct {
	Items []PlaylistItemUpdateInfo `json:"items" binding:"required,dive"`
}

type RemovePlaylistItemsRequest struct {
	IDs []int64 `json:"ids" binding:"required"`
}

type GetReactionsRequest struct {
	ListRequest
	UIDsFilter
}

type ReactionsResponse struct {
	ListResponse
	Items []*Reaction `json:"items"`
}

type AddReactionsRequest struct {
	Kind        string `json:"kind" binding:"required,max=16"`
	SubjectType string `json:"subject_type" binding:"required,max=16"`
	SubjectUID  string `json:"subject_uid" binding:"required,len=8"`
}

type RemoveReactionsRequest struct {
	AddReactionsRequest
}

type GetHistoryRequest struct {
	ListRequest
}

type HistoryResponse struct {
	ListResponse
	Items []*History `json:"items"`
}

type GetSubscriptionsRequest struct {
	ListRequest
}

type SubscriptionsResponse struct {
	ListResponse
	Items []*Subscription `json:"items"`
}

type SubscribeRequest struct {
	CollectionUID  string `json:"collection_uid" binding:"omitempty"`
	ContentType    string `json:"content_type" binding:"omitempty"`
	ContentUnitUID string `json:"content_unit_uid" binding:"omitempty"`
}

// DTOs

type Playlist struct {
	ID         int64                  `json:"id"`
	UID        string                 `json:"uid"`
	UserID     int64                  `json:"user_id"`
	Name       string                 `json:"name,omitempty"`
	Public     bool                   `json:"public"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	TotalItems int                    `json:"total_items"`
	Items      []*PlaylistItem        `json:"items"`
}

type PlaylistItem struct {
	ID             int64  `json:"id"`
	Position       int    `json:"position"`
	ContentUnitUID string `json:"content_unit_uid"`
}

type Reaction struct {
	Kind        string `json:"kind"`
	SubjectType string `json:"subject_type"`
	SubjectUID  string `json:"subject_uid"`
}

type History struct {
	ID             int64       `json:"id"`
	ContentUnitUID null.String `json:"content_unit_uid,omitempty"`
	Data           null.JSON   `json:"data,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
}

type Subscription struct {
	ID             int64       `json:"id"`
	CollectionUID  null.String `json:"collection_uid,omitempty"`
	ContentType    null.String `json:"content_type,omitempty"`
	ContentUnitUID null.String `json:"content_unit_uid,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      null.Time   `json:"updated_at,omitempty"`
}
