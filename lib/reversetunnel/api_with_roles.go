/*
Copyright 2020 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package reversetunnel

import (
	"context"

	"github.com/gravitational/teleport/lib/auth"
	"github.com/gravitational/teleport/lib/services"

	"github.com/gravitational/trace"
)

// NewServerWithRoles returns new authorizing server
func NewServerWithRoles(server Server, roles services.RoleSet, ap auth.AccessPoint) *ServerWithRoles {
	return &ServerWithRoles{
		server: server,
		roles:  roles,
		ap:     ap,
	}
}

// ServerWithRoles authorizes requests
type ServerWithRoles struct {
	server Server
	// roles is a set of roles used to check RBAC permissions.
	roles services.RoleSet
	// ap is access point
	ap auth.AccessPoint
}

// GetSites returns a list of connected remote sites
func (s *ServerWithRoles) GetSites() ([]RemoteSite, error) {
	clusters, err := s.server.GetSites()
	if err != nil {
		return nil, trace.Wrap(err)
	}
	out := make([]RemoteSite, 0, len(clusters))
	for _, cluster := range clusters {
		if _, ok := cluster.(*localSite); ok {
			out = append(out, cluster)
			continue
		}
		rc, err := s.ap.GetRemoteCluster(cluster.GetName())
		if err != nil {
			return nil, trace.Wrap(err)
		}
		if err := s.roles.CheckAccessToRemoteCluster(rc); err != nil {
			if !trace.IsAccessDenied(err) {
				return nil, trace.Wrap(err)
			}
			continue
		}
		out = append(out, cluster)
	}
	return out, nil
}

// GetSite returns remote site this node belongs to
func (s *ServerWithRoles) GetSite(clusterName string) (RemoteSite, error) {
	cluster, err := s.server.GetSite(clusterName)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	if _, ok := cluster.(*localSite); ok {
		return cluster, nil
	}
	rc, err := s.ap.GetRemoteCluster(clusterName)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	if err := s.roles.CheckAccessToRemoteCluster(rc); err != nil {
		return nil, trace.Wrap(err)
	}
	return cluster, nil
}

// Start starts server
func (s *ServerWithRoles) Start() error {
	return s.server.Start()
}

// Close closes server's operations immediately
func (s *ServerWithRoles) Close() error {
	return s.server.Close()
}

// Shutdown performs graceful server shutdown
func (s *ServerWithRoles) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// Wait waits for server to close all outstanding operations
func (s *ServerWithRoles) Wait() {
	s.server.Wait()
}
