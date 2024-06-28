// Copyright 2021 Nitric Pty Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package postgresql

import (
	"context"
	"fmt"
	"os"

	sqlpb "github.com/nitrictech/nitric/core/pkg/proto/sql/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SQLDatabaseService - Nitric Secret Service implementation for SQL Database
type PostgresSqlService struct{}

var _ sqlpb.SqlServer = &PostgresSqlService{}

func (s *PostgresSqlService) ConnectionString(ctx context.Context, req *sqlpb.SqlConnectionStringRequest) (*sqlpb.SqlConnectionStringResponse, error) {
	baseUrl := os.Getenv("NITRIC_DATABASE_BASE_URL")

	if baseUrl == "" {
		return nil, status.Error(codes.FailedPrecondition, "NITRIC_DATABASE_BASE_URL environment variable not set")
	}

	return &sqlpb.SqlConnectionStringResponse{
		ConnectionString: fmt.Sprintf("%s/%s?sslmode=require", baseUrl, req.DatabaseName),
	}, nil
}

// New - Creates a new Nitric SQL service with Azure PostgreSQL Provider
func New() (*PostgresSqlService, error) {
	return &PostgresSqlService{}, nil
}
