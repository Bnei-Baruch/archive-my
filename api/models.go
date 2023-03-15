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

type NameFilter struct {
	Name string `json:"name" binding:"omitempty,max=256"`
}

//Filters
type IDsFilter struct {
	IDs []int64 `json:"ids" form:"ids" binding:"omitempty"`
}

type UIDsFilter struct {
	UIDs []string `json:"uids" form:"uids" binding:"omitempty,dive,len=8"`
}

type CUUIDsFilter struct {
	CUUIDs []string `json:"cu_uids" form:"cu_uids" binding:"omitempty,dive,len=8"`
}

type QueryFilter struct {
	Query string `json:"query" form:"query" binding:"omitempty"`
}

//Playlist
type GetPlaylistsRequest struct {
	ListRequest
	ExistUnit string `form:"exist_cu" binding:"omitempty"`
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
	NameFilter
	Public     bool                   `json:"public"`
	Properties map[string]interface{} `json:"properties" binding:"omitempty"`
}

//PlaylistItem
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

//Reaction
type GetReactionsRequest struct {
	ListRequest
	UIDsFilter
	SubjectType string `json:"subject_type" form:"subject_type" binding:"omitempty"`
}

type ReactionsResponse struct {
	ListResponse
	Items []*Reaction `json:"items"`
}

type AddReactionsRequest struct {
	Kind        string `json:"kind" binding:"required,max=16"`
	SubjectType string `json:"subject_type" binding:"required,max=32"`
	SubjectUID  string `json:"subject_uid" binding:"required,len=8"`
}

type RemoveReactionsRequest struct {
	AddReactionsRequest
}

type ReactionCountRequest struct {
	UIDsFilter
	SubjectType string `form:"type" binding:"omitempty"`
}

//History
type GetHistoryRequest struct {
	ListRequest
	CUUIDsFilter
}

type HistoryResponse struct {
	ListResponse
	Items []*History `json:"items"`
}

//Subscription
type GetSubscriptionsRequest struct {
	ListRequest
	CollectionUID  string `form:"collection_uid" binding:"omitempty"`
	ContentType    string `form:"content_type" binding:"omitempty"`
	ContentUnitUID string `form:"content_unit_uid" binding:"omitempty"`
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

//Bookmark
type GetBookmarksRequest struct {
	ListRequest
	QueryFilter
	FolderIDsFilter []int64 `form:"folder_id" binding:"omitempty"`
}

type GetBookmarksResponse struct {
	ListResponse
	Items []*Bookmark `json:"items"`
}

type AddBookmarksRequest struct {
	NameFilter
	SubjectUID  string                 `json:"subject_uid" binding:"required,max=8"`
	SubjectType string                 `json:"subject_type" binding:"required"`
	FolderIDs   []int64                `json:"folder_ids" binding:"omitempty"`
	Properties  map[string]interface{} `json:"properties" binding:"omitempty"`
}

type UpdateBookmarkRequest struct {
	NameFilter
	FolderIDs  []int64                `json:"folder_ids" binding:"omitempty"`
	Properties map[string]interface{} `json:"properties" binding:"omitempty"`
}

//Folder
type GetFoldersRequest struct {
	ListRequest
	QueryFilter
	BookmarkIdFilter int64 `form:"bookmark_id" binding:"omitempty"`
}

type GetFoldersResponse struct {
	ListResponse
	Items []*Folder `json:"items"`
}

type AddFolderRequest struct {
	NameFilter
}

type UpdateFolderRequest struct {
	NameFilter
}

// DTOs

type Playlist struct {
	ID            int64                  `json:"id"`
	UID           string                 `json:"uid"`
	Name          string                 `json:"name,omitempty"`
	Public        bool                   `json:"public"`
	Properties    map[string]interface{} `json:"properties,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	TotalItems    int                    `json:"total_items"`
	Items         []*PlaylistItem        `json:"items"`
	PosterUnitUID string                 `json:"poster_unit_uid,omitempty"`
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

type ReactionCount struct {
	Reaction
	Total int `json:"total"`
}

type History struct {
	ID             int64       `json:"id"`
	ContentUnitUID null.String `json:"content_unit_uid,omitempty"`
	Data           null.JSON   `json:"data,omitempty"`
	Timestamp      time.Time   `json:"timestamp"`
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

type Bookmark struct {
	ID          int64                  `json:"id"`
	Name        string                 `json:"name"`
	SubjectUID  string                 `json:"subject_uid"`
	SubjectType string                 `json:"subject_type"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	FolderIds   []int64                `json:"folder_ids,omitempty"`
}

type Folder struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"created_at"`
	BookmarkIds []int64   `json:"bookmark_ids,omitempty"`
}

// Languages
type Languages struct {
	Languages      []string `json:"languages"`
	ShowAllContent bool     `json:"show_all_content"`
}
