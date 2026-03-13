package mcpserver

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mreyeswilson/pocketmcp/internal/pocketbase"
)

type Server struct {
	pb *pocketbase.Client
}

type ListCollectionsInput struct {
	Page    int    `json:"page,omitempty" jsonschema:"page number"`
	PerPage int    `json:"perPage,omitempty" jsonschema:"maximum number of collections to return"`
	Sort    string `json:"sort,omitempty" jsonschema:"sort expression"`
	Filter  string `json:"filter,omitempty" jsonschema:"PocketBase filter expression"`
	Fields  string `json:"fields,omitempty" jsonschema:"comma-separated fields projection"`
}

type GetCollectionInput struct {
	Collection string `json:"collection" jsonschema:"collection id or name"`
	Fields     string `json:"fields,omitempty" jsonschema:"comma-separated fields projection"`
}

type CreateCollectionInput struct {
	Body map[string]any `json:"body" jsonschema:"PocketBase collection payload"`
}

type UpdateCollectionInput struct {
	Collection string         `json:"collection" jsonschema:"collection id or name"`
	Body       map[string]any `json:"body" jsonschema:"partial collection payload"`
}

type DeleteCollectionInput struct {
	Collection string `json:"collection" jsonschema:"collection id or name"`
}

type ListRecordsInput struct {
	Collection string `json:"collection" jsonschema:"collection id or name, including auth collections for users"`
	Page       int    `json:"page,omitempty" jsonschema:"page number"`
	PerPage    int    `json:"perPage,omitempty" jsonschema:"maximum number of records to return"`
	Sort       string `json:"sort,omitempty" jsonschema:"sort expression"`
	Filter     string `json:"filter,omitempty" jsonschema:"PocketBase filter expression"`
	Expand     string `json:"expand,omitempty" jsonschema:"relations to expand"`
	Fields     string `json:"fields,omitempty" jsonschema:"comma-separated fields projection"`
	SkipTotal  bool   `json:"skipTotal,omitempty" jsonschema:"skip total counter calculation"`
}

type GetRecordInput struct {
	Collection string `json:"collection" jsonschema:"collection id or name, including auth collections for users"`
	RecordID   string `json:"recordId" jsonschema:"record id"`
	Expand     string `json:"expand,omitempty" jsonschema:"relations to expand"`
	Fields     string `json:"fields,omitempty" jsonschema:"comma-separated fields projection"`
}

type CreateRecordInput struct {
	Collection string         `json:"collection" jsonschema:"collection id or name, including auth collections for users"`
	Body       map[string]any `json:"body" jsonschema:"record payload; auth collections can include password and passwordConfirm"`
	Expand     string         `json:"expand,omitempty" jsonschema:"relations to expand"`
	Fields     string         `json:"fields,omitempty" jsonschema:"comma-separated fields projection"`
}

type UpdateRecordInput struct {
	Collection string         `json:"collection" jsonschema:"collection id or name, including auth collections for users"`
	RecordID   string         `json:"recordId" jsonschema:"record id"`
	Body       map[string]any `json:"body" jsonschema:"partial record payload; auth collections can include password changes"`
	Expand     string         `json:"expand,omitempty" jsonschema:"relations to expand"`
	Fields     string         `json:"fields,omitempty" jsonschema:"comma-separated fields projection"`
}

type DeleteRecordInput struct {
	Collection string `json:"collection" jsonschema:"collection id or name, including auth collections for users"`
	RecordID   string `json:"recordId" jsonschema:"record id"`
}

type GetSettingsInput struct {
	Fields string `json:"fields,omitempty" jsonschema:"comma-separated fields projection"`
}

type UpdateSettingsInput struct {
	Body map[string]any `json:"body" jsonschema:"PocketBase settings payload"`
}

func New(pb *pocketbase.Client, version string) *mcp.Server {
	srv := &Server{pb: pb}
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "pocketbase-admin",
		Version: version,
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_collections",
		Description: "List PocketBase collections as superuser. Use this before table/schema administration.",
	}, srv.listCollections)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_collection",
		Description: "Get a single PocketBase collection by id or name.",
	}, srv.getCollection)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_collection",
		Description: "Create a PocketBase collection. Supports base, auth, and view collections.",
	}, srv.createCollection)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_collection",
		Description: "Update a PocketBase collection schema, API rules, indexes, or auth/view options.",
	}, srv.updateCollection)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_collection",
		Description: "Delete a PocketBase collection by id or name.",
	}, srv.deleteCollection)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_records",
		Description: "List records from any PocketBase collection, including auth collections for users.",
	}, srv.listRecords)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_record",
		Description: "Get a single PocketBase record by collection and record id.",
	}, srv.getRecord)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_record",
		Description: "Create a record in any PocketBase collection, including auth collections for users.",
	}, srv.createRecord)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_record",
		Description: "Update a record in any PocketBase collection, including auth collections for users.",
	}, srv.updateRecord)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_record",
		Description: "Delete a record in any PocketBase collection, including auth collections for users.",
	}, srv.deleteRecord)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_settings",
		Description: "Read PocketBase application settings as superuser.",
	}, srv.getSettings)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_settings",
		Description: "Update PocketBase application settings as superuser.",
	}, srv.updateSettings)

	return server
}

func (s *Server) listCollections(ctx context.Context, _ *mcp.CallToolRequest, in ListCollectionsInput) (*mcp.CallToolResult, map[string]any, error) {
	result, err := s.pb.ListCollections(ctx, buildPaginationParams(in.Page, in.PerPage, in.Sort, in.Filter, "", in.Fields, false))
	return nil, result, err
}

func (s *Server) getCollection(ctx context.Context, _ *mcp.CallToolRequest, in GetCollectionInput) (*mcp.CallToolResult, map[string]any, error) {
	result, err := s.pb.GetCollection(ctx, in.Collection, buildFieldsParams(in.Fields))
	return nil, result, err
}

func (s *Server) createCollection(ctx context.Context, _ *mcp.CallToolRequest, in CreateCollectionInput) (*mcp.CallToolResult, map[string]any, error) {
	result, err := s.pb.CreateCollection(ctx, in.Body)
	return nil, result, err
}

func (s *Server) updateCollection(ctx context.Context, _ *mcp.CallToolRequest, in UpdateCollectionInput) (*mcp.CallToolResult, map[string]any, error) {
	result, err := s.pb.UpdateCollection(ctx, in.Collection, in.Body)
	return nil, result, err
}

func (s *Server) deleteCollection(ctx context.Context, _ *mcp.CallToolRequest, in DeleteCollectionInput) (*mcp.CallToolResult, map[string]any, error) {
	result, err := s.pb.DeleteCollection(ctx, in.Collection)
	return nil, result, err
}

func (s *Server) listRecords(ctx context.Context, _ *mcp.CallToolRequest, in ListRecordsInput) (*mcp.CallToolResult, map[string]any, error) {
	params := buildPaginationParams(in.Page, in.PerPage, in.Sort, in.Filter, in.Expand, in.Fields, in.SkipTotal)
	result, err := s.pb.ListRecords(ctx, in.Collection, params)
	return nil, result, err
}

func (s *Server) getRecord(ctx context.Context, _ *mcp.CallToolRequest, in GetRecordInput) (*mcp.CallToolResult, map[string]any, error) {
	result, err := s.pb.GetRecord(ctx, in.Collection, in.RecordID, buildRecordProjectionParams(in.Expand, in.Fields))
	return nil, result, err
}

func (s *Server) createRecord(ctx context.Context, _ *mcp.CallToolRequest, in CreateRecordInput) (*mcp.CallToolResult, map[string]any, error) {
	result, err := s.pb.CreateRecord(ctx, in.Collection, in.Body, buildRecordProjectionParams(in.Expand, in.Fields))
	return nil, result, err
}

func (s *Server) updateRecord(ctx context.Context, _ *mcp.CallToolRequest, in UpdateRecordInput) (*mcp.CallToolResult, map[string]any, error) {
	result, err := s.pb.UpdateRecord(ctx, in.Collection, in.RecordID, in.Body, buildRecordProjectionParams(in.Expand, in.Fields))
	return nil, result, err
}

func (s *Server) deleteRecord(ctx context.Context, _ *mcp.CallToolRequest, in DeleteRecordInput) (*mcp.CallToolResult, map[string]any, error) {
	result, err := s.pb.DeleteRecord(ctx, in.Collection, in.RecordID)
	return nil, result, err
}

func (s *Server) getSettings(ctx context.Context, _ *mcp.CallToolRequest, in GetSettingsInput) (*mcp.CallToolResult, map[string]any, error) {
	result, err := s.pb.GetSettings(ctx, buildFieldsParams(in.Fields))
	return nil, result, err
}

func (s *Server) updateSettings(ctx context.Context, _ *mcp.CallToolRequest, in UpdateSettingsInput) (*mcp.CallToolResult, map[string]any, error) {
	result, err := s.pb.UpdateSettings(ctx, in.Body)
	return nil, result, err
}

func buildPaginationParams(page int, perPage int, sort string, filter string, expand string, fields string, skipTotal bool) url.Values {
	values := url.Values{}
	if page > 0 {
		values.Set("page", strconv.Itoa(page))
	}
	if perPage > 0 {
		values.Set("perPage", strconv.Itoa(perPage))
	}
	setIfNotEmpty(values, "sort", sort)
	setIfNotEmpty(values, "filter", filter)
	setIfNotEmpty(values, "expand", expand)
	setIfNotEmpty(values, "fields", fields)
	if skipTotal {
		values.Set("skipTotal", "true")
	}
	return values
}

func buildFieldsParams(fields string) url.Values {
	values := url.Values{}
	setIfNotEmpty(values, "fields", fields)
	return values
}

func buildRecordProjectionParams(expand string, fields string) url.Values {
	values := url.Values{}
	setIfNotEmpty(values, "expand", expand)
	setIfNotEmpty(values, "fields", fields)
	return values
}

func setIfNotEmpty(values url.Values, key string, value string) {
	if trimmed := strings.TrimSpace(value); trimmed != "" {
		values.Set(key, trimmed)
	}
}
