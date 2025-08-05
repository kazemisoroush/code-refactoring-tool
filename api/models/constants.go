// Package models provides constants for authentication and authorization
package models

// UserRole defines the role of a user in the system
type UserRole string

const (
	// RoleOwner has full access to everything
	RoleOwner UserRole = "owner" // Full access to everything
	// RoleAdmin can manage users and all projects
	RoleAdmin UserRole = "admin" // Manage users, all projects
	// RoleDeveloper can create and manage own projects
	RoleDeveloper UserRole = "developer" // Create/manage own projects
	// RoleViewer has read-only access
	RoleViewer UserRole = "viewer" // Read-only access
)

// UserStatus represents the current status of a user account
type UserStatus string

const (
	// UserStatusActive indicates an active user account
	UserStatusActive UserStatus = "active"
	// UserStatusInactive indicates an inactive user account
	UserStatusInactive UserStatus = "inactive"
	// UserStatusPending indicates a user account pending activation
	UserStatusPending UserStatus = "pending"
	// UserStatusSuspended indicates a suspended user account
	UserStatusSuspended UserStatus = "suspended"
)

// Permission represents a specific permission in the system
type Permission string

const (
	// PermissionProjectCreate allows creating new projects
	PermissionProjectCreate Permission = "project:create"
	// PermissionProjectRead allows reading project information
	PermissionProjectRead Permission = "project:read"
	// PermissionProjectUpdate allows updating project information
	PermissionProjectUpdate Permission = "project:update"
	// PermissionProjectDelete allows deleting projects
	PermissionProjectDelete Permission = "project:delete"

	// PermissionTaskCreate allows creating new tasks
	PermissionTaskCreate Permission = "task:create"
	// PermissionTaskExecute allows executing tasks
	PermissionTaskExecute Permission = "task:execute"
	// PermissionTaskRead allows reading task information
	PermissionTaskRead Permission = "task:read"
	// PermissionTaskUpdate allows updating task information
	PermissionTaskUpdate Permission = "task:update"
	// PermissionTaskDelete allows deleting tasks
	PermissionTaskDelete Permission = "task:delete"

	// PermissionAgentCreate allows creating new agents
	PermissionAgentCreate Permission = "agent:create"
	// PermissionAgentRead allows reading agent information
	PermissionAgentRead Permission = "agent:read"
	// PermissionAgentUpdate allows updating agent information
	PermissionAgentUpdate Permission = "agent:update"
	// PermissionAgentDelete allows deleting agents
	PermissionAgentDelete Permission = "agent:delete"

	// PermissionUserCreate allows creating new users
	PermissionUserCreate Permission = "user:create"
	// PermissionUserRead allows reading user information
	PermissionUserRead Permission = "user:read"
	// PermissionUserUpdate allows updating user information
	PermissionUserUpdate Permission = "user:update"
	// PermissionUserDelete allows deleting users
	PermissionUserDelete Permission = "user:delete"
)

// AgentStatus represents the current status of an agent
type AgentStatus string

// Agent status constants define the possible states of an agent
const (
	// AgentStatusPending indicates the agent is waiting to be initialized
	AgentStatusPending AgentStatus = "pending"
	// AgentStatusInitializing indicates the agent is being set up
	AgentStatusInitializing AgentStatus = "initializing"
	// AgentStatusReady indicates the agent is ready to process tasks
	AgentStatusReady AgentStatus = "ready"
	// AgentStatusFailed indicates the agent failed to initialize or operate
	AgentStatusFailed AgentStatus = "failed"
)
