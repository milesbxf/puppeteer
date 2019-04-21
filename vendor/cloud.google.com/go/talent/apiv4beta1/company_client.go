// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by gapic-generator. DO NOT EDIT.

package talent

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/golang/protobuf/proto"
	gax "github.com/googleapis/gax-go/v2"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/api/transport"
	talentpb "google.golang.org/genproto/googleapis/cloud/talent/v4beta1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

// CompanyCallOptions contains the retry settings for each method of CompanyClient.
type CompanyCallOptions struct {
	CreateCompany []gax.CallOption
	GetCompany    []gax.CallOption
	UpdateCompany []gax.CallOption
	DeleteCompany []gax.CallOption
	ListCompanies []gax.CallOption
}

func defaultCompanyClientOptions() []option.ClientOption {
	return []option.ClientOption{
		option.WithEndpoint("jobs.googleapis.com:443"),
		option.WithScopes(DefaultAuthScopes()...),
	}
}

func defaultCompanyCallOptions() *CompanyCallOptions {
	retry := map[[2]string][]gax.CallOption{
		{"default", "idempotent"}: {
			gax.WithRetry(func() gax.Retryer {
				return gax.OnCodes([]codes.Code{
					codes.DeadlineExceeded,
					codes.Unavailable,
				}, gax.Backoff{
					Initial:    100 * time.Millisecond,
					Max:        60000 * time.Millisecond,
					Multiplier: 1.3,
				})
			}),
		},
	}
	return &CompanyCallOptions{
		CreateCompany: retry[[2]string{"default", "non_idempotent"}],
		GetCompany:    retry[[2]string{"default", "idempotent"}],
		UpdateCompany: retry[[2]string{"default", "non_idempotent"}],
		DeleteCompany: retry[[2]string{"default", "idempotent"}],
		ListCompanies: retry[[2]string{"default", "idempotent"}],
	}
}

// CompanyClient is a client for interacting with Cloud Talent Solution API.
//
// Methods, except Close, may be called concurrently. However, fields must not be modified concurrently with method calls.
type CompanyClient struct {
	// The connection to the service.
	conn *grpc.ClientConn

	// The gRPC API client.
	companyClient talentpb.CompanyServiceClient

	// The call options for this service.
	CallOptions *CompanyCallOptions

	// The x-goog-* metadata to be sent with each request.
	xGoogMetadata metadata.MD
}

// NewCompanyClient creates a new company service client.
//
// A service that handles company management, including CRUD and enumeration.
func NewCompanyClient(ctx context.Context, opts ...option.ClientOption) (*CompanyClient, error) {
	conn, err := transport.DialGRPC(ctx, append(defaultCompanyClientOptions(), opts...)...)
	if err != nil {
		return nil, err
	}
	c := &CompanyClient{
		conn:        conn,
		CallOptions: defaultCompanyCallOptions(),

		companyClient: talentpb.NewCompanyServiceClient(conn),
	}
	c.setGoogleClientInfo()
	return c, nil
}

// Connection returns the client's connection to the API service.
func (c *CompanyClient) Connection() *grpc.ClientConn {
	return c.conn
}

// Close closes the connection to the API service. The user should invoke this when
// the client is no longer required.
func (c *CompanyClient) Close() error {
	return c.conn.Close()
}

// setGoogleClientInfo sets the name and version of the application in
// the `x-goog-api-client` header passed on each request. Intended for
// use by Google-written clients.
func (c *CompanyClient) setGoogleClientInfo(keyval ...string) {
	kv := append([]string{"gl-go", versionGo()}, keyval...)
	kv = append(kv, "gapic", versionClient, "gax", gax.Version, "grpc", grpc.Version)
	c.xGoogMetadata = metadata.Pairs("x-goog-api-client", gax.XGoogHeader(kv...))
}

// CreateCompany creates a new company entity.
func (c *CompanyClient) CreateCompany(ctx context.Context, req *talentpb.CreateCompanyRequest, opts ...gax.CallOption) (*talentpb.Company, error) {
	md := metadata.Pairs("x-goog-request-params", fmt.Sprintf("%s=%v", "parent", req.GetParent()))
	ctx = insertMetadata(ctx, c.xGoogMetadata, md)
	opts = append(c.CallOptions.CreateCompany[0:len(c.CallOptions.CreateCompany):len(c.CallOptions.CreateCompany)], opts...)
	var resp *talentpb.Company
	err := gax.Invoke(ctx, func(ctx context.Context, settings gax.CallSettings) error {
		var err error
		resp, err = c.companyClient.CreateCompany(ctx, req, settings.GRPC...)
		return err
	}, opts...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetCompany retrieves specified company.
func (c *CompanyClient) GetCompany(ctx context.Context, req *talentpb.GetCompanyRequest, opts ...gax.CallOption) (*talentpb.Company, error) {
	md := metadata.Pairs("x-goog-request-params", fmt.Sprintf("%s=%v", "name", req.GetName()))
	ctx = insertMetadata(ctx, c.xGoogMetadata, md)
	opts = append(c.CallOptions.GetCompany[0:len(c.CallOptions.GetCompany):len(c.CallOptions.GetCompany)], opts...)
	var resp *talentpb.Company
	err := gax.Invoke(ctx, func(ctx context.Context, settings gax.CallSettings) error {
		var err error
		resp, err = c.companyClient.GetCompany(ctx, req, settings.GRPC...)
		return err
	}, opts...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// UpdateCompany updates specified company.
func (c *CompanyClient) UpdateCompany(ctx context.Context, req *talentpb.UpdateCompanyRequest, opts ...gax.CallOption) (*talentpb.Company, error) {
	md := metadata.Pairs("x-goog-request-params", fmt.Sprintf("%s=%v", "company.name", req.GetCompany().GetName()))
	ctx = insertMetadata(ctx, c.xGoogMetadata, md)
	opts = append(c.CallOptions.UpdateCompany[0:len(c.CallOptions.UpdateCompany):len(c.CallOptions.UpdateCompany)], opts...)
	var resp *talentpb.Company
	err := gax.Invoke(ctx, func(ctx context.Context, settings gax.CallSettings) error {
		var err error
		resp, err = c.companyClient.UpdateCompany(ctx, req, settings.GRPC...)
		return err
	}, opts...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// DeleteCompany deletes specified company.
// Prerequisite: The company has no jobs associated with it.
func (c *CompanyClient) DeleteCompany(ctx context.Context, req *talentpb.DeleteCompanyRequest, opts ...gax.CallOption) error {
	md := metadata.Pairs("x-goog-request-params", fmt.Sprintf("%s=%v", "name", req.GetName()))
	ctx = insertMetadata(ctx, c.xGoogMetadata, md)
	opts = append(c.CallOptions.DeleteCompany[0:len(c.CallOptions.DeleteCompany):len(c.CallOptions.DeleteCompany)], opts...)
	err := gax.Invoke(ctx, func(ctx context.Context, settings gax.CallSettings) error {
		var err error
		_, err = c.companyClient.DeleteCompany(ctx, req, settings.GRPC...)
		return err
	}, opts...)
	return err
}

// ListCompanies lists all companies associated with the project.
func (c *CompanyClient) ListCompanies(ctx context.Context, req *talentpb.ListCompaniesRequest, opts ...gax.CallOption) *CompanyIterator {
	md := metadata.Pairs("x-goog-request-params", fmt.Sprintf("%s=%v", "parent", req.GetParent()))
	ctx = insertMetadata(ctx, c.xGoogMetadata, md)
	opts = append(c.CallOptions.ListCompanies[0:len(c.CallOptions.ListCompanies):len(c.CallOptions.ListCompanies)], opts...)
	it := &CompanyIterator{}
	req = proto.Clone(req).(*talentpb.ListCompaniesRequest)
	it.InternalFetch = func(pageSize int, pageToken string) ([]*talentpb.Company, string, error) {
		var resp *talentpb.ListCompaniesResponse
		req.PageToken = pageToken
		if pageSize > math.MaxInt32 {
			req.PageSize = math.MaxInt32
		} else {
			req.PageSize = int32(pageSize)
		}
		err := gax.Invoke(ctx, func(ctx context.Context, settings gax.CallSettings) error {
			var err error
			resp, err = c.companyClient.ListCompanies(ctx, req, settings.GRPC...)
			return err
		}, opts...)
		if err != nil {
			return nil, "", err
		}
		return resp.Companies, resp.NextPageToken, nil
	}
	fetch := func(pageSize int, pageToken string) (string, error) {
		items, nextPageToken, err := it.InternalFetch(pageSize, pageToken)
		if err != nil {
			return "", err
		}
		it.items = append(it.items, items...)
		return nextPageToken, nil
	}
	it.pageInfo, it.nextFunc = iterator.NewPageInfo(fetch, it.bufLen, it.takeBuf)
	it.pageInfo.MaxSize = int(req.PageSize)
	return it
}

// CompanyIterator manages a stream of *talentpb.Company.
type CompanyIterator struct {
	items    []*talentpb.Company
	pageInfo *iterator.PageInfo
	nextFunc func() error

	// InternalFetch is for use by the Google Cloud Libraries only.
	// It is not part of the stable interface of this package.
	//
	// InternalFetch returns results from a single call to the underlying RPC.
	// The number of results is no greater than pageSize.
	// If there are no more results, nextPageToken is empty and err is nil.
	InternalFetch func(pageSize int, pageToken string) (results []*talentpb.Company, nextPageToken string, err error)
}

// PageInfo supports pagination. See the google.golang.org/api/iterator package for details.
func (it *CompanyIterator) PageInfo() *iterator.PageInfo {
	return it.pageInfo
}

// Next returns the next result. Its second return value is iterator.Done if there are no more
// results. Once Next returns Done, all subsequent calls will return Done.
func (it *CompanyIterator) Next() (*talentpb.Company, error) {
	var item *talentpb.Company
	if err := it.nextFunc(); err != nil {
		return item, err
	}
	item = it.items[0]
	it.items = it.items[1:]
	return item, nil
}

func (it *CompanyIterator) bufLen() int {
	return len(it.items)
}

func (it *CompanyIterator) takeBuf() interface{} {
	b := it.items
	it.items = nil
	return b
}
