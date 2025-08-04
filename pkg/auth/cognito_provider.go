package auth

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

// CognitoConfig contains configuration for Cognito authentication
type CognitoConfig struct {
	UserPoolID   string
	ClientID     string
	ClientSecret *string // Optional, for confidential clients
	Region       string
}

// CognitoProvider implements AuthProvider using AWS Cognito
type CognitoProvider struct {
	client *cognitoidentityprovider.Client
	config CognitoConfig
}

// NewCognitoProvider creates a new Cognito authentication provider
func NewCognitoProvider(client *cognitoidentityprovider.Client, config CognitoConfig) *CognitoProvider {
	return &CognitoProvider{
		client: client,
		config: config,
	}
}

// CreateUser creates a new user in Cognito
func (c *CognitoProvider) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
	input := &cognitoidentityprovider.AdminCreateUserInput{
		UserPoolId: aws.String(c.config.UserPoolID),
		Username:   aws.String(req.Username),
		UserAttributes: []types.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(req.Email),
			},
			{
				Name:  aws.String("email_verified"),
				Value: aws.String("true"),
			},
		},
		TemporaryPassword: req.Password,
		MessageAction:     types.MessageActionTypeSuppress,
	}

	if req.FirstName != nil && *req.FirstName != "" {
		input.UserAttributes = append(input.UserAttributes, types.AttributeType{
			Name:  aws.String("given_name"),
			Value: req.FirstName,
		})
	}

	if req.LastName != nil && *req.LastName != "" {
		input.UserAttributes = append(input.UserAttributes, types.AttributeType{
			Name:  aws.String("family_name"),
			Value: req.LastName,
		})
	}

	result, err := c.client.AdminCreateUser(ctx, input)
	if err != nil {
		return nil, c.mapCognitoError(err)
	}

	return c.mapCognitoUserToUser(result.User), nil
}

// GetUser retrieves a user by ID from Cognito
func (c *CognitoProvider) GetUser(ctx context.Context, userID string) (*User, error) {
	input := &cognitoidentityprovider.AdminGetUserInput{
		UserPoolId: aws.String(c.config.UserPoolID),
		Username:   aws.String(userID),
	}

	result, err := c.client.AdminGetUser(ctx, input)
	if err != nil {
		return nil, c.mapCognitoError(err)
	}

	user := &User{
		ID:        userID,
		Username:  userID,
		Status:    c.mapCognitoUserStatus(result.UserStatus),
		CreatedAt: aws.ToTime(result.UserCreateDate),
		UpdatedAt: aws.ToTime(result.UserLastModifiedDate),
	}

	// Extract attributes
	for _, attr := range result.UserAttributes {
		switch aws.ToString(attr.Name) {
		case "email":
			user.Email = aws.ToString(attr.Value)
		case "given_name":
			user.FirstName = attr.Value
		case "family_name":
			user.LastName = attr.Value
		}
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email from Cognito
func (c *CognitoProvider) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	input := &cognitoidentityprovider.ListUsersInput{
		UserPoolId: aws.String(c.config.UserPoolID),
		Filter:     aws.String("email = \"" + email + "\""),
		Limit:      aws.Int32(1),
	}

	result, err := c.client.ListUsers(ctx, input)
	if err != nil {
		return nil, c.mapCognitoError(err)
	}

	if len(result.Users) == 0 {
		return nil, ErrUserNotFound
	}

	return c.mapCognitoUserToUser(&result.Users[0]), nil
}

// UpdateUser updates a user in Cognito
func (c *CognitoProvider) UpdateUser(ctx context.Context, userID string, req *UpdateUserRequest) (*User, error) {
	var attributes []types.AttributeType

	if req.Email != nil && *req.Email != "" {
		attributes = append(attributes, types.AttributeType{
			Name:  aws.String("email"),
			Value: req.Email,
		})
	}

	if req.FirstName != nil && *req.FirstName != "" {
		attributes = append(attributes, types.AttributeType{
			Name:  aws.String("given_name"),
			Value: req.FirstName,
		})
	}

	if req.LastName != nil && *req.LastName != "" {
		attributes = append(attributes, types.AttributeType{
			Name:  aws.String("family_name"),
			Value: req.LastName,
		})
	}

	if len(attributes) > 0 {
		input := &cognitoidentityprovider.AdminUpdateUserAttributesInput{
			UserPoolId:     aws.String(c.config.UserPoolID),
			Username:       aws.String(userID),
			UserAttributes: attributes,
		}

		_, err := c.client.AdminUpdateUserAttributes(ctx, input)
		if err != nil {
			return nil, c.mapCognitoError(err)
		}
	}

	return c.GetUser(ctx, userID)
}

// DeleteUser deletes a user from Cognito
func (c *CognitoProvider) DeleteUser(ctx context.Context, userID string) error {
	input := &cognitoidentityprovider.AdminDeleteUserInput{
		UserPoolId: aws.String(c.config.UserPoolID),
		Username:   aws.String(userID),
	}

	_, err := c.client.AdminDeleteUser(ctx, input)
	return c.mapCognitoError(err)
}

// ListUsers lists users from Cognito
func (c *CognitoProvider) ListUsers(ctx context.Context, req *ListUsersRequest) (*ListUsersResponse, error) {
	input := &cognitoidentityprovider.ListUsersInput{
		UserPoolId: aws.String(c.config.UserPoolID),
	}

	if req.Limit > 0 {
		input.Limit = aws.Int32(int32(req.Limit))
	}

	result, err := c.client.ListUsers(ctx, input)
	if err != nil {
		return nil, c.mapCognitoError(err)
	}

	users := make([]*User, len(result.Users))
	for i, cognitoUser := range result.Users {
		users[i] = c.mapCognitoUserToUser(&cognitoUser)
	}

	response := &ListUsersResponse{
		Users: users,
	}

	if result.PaginationToken != nil {
		response.NextToken = result.PaginationToken
	}

	return response, nil
}

// SignUp creates a new user account in Cognito
func (c *CognitoProvider) SignUp(ctx context.Context, req *SignUpRequest) (*AuthResult, error) {
	input := &cognitoidentityprovider.SignUpInput{
		ClientId: aws.String(c.config.ClientID),
		Username: aws.String(req.Email),
		Password: aws.String(req.Password),
		UserAttributes: []types.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(req.Email),
			},
		},
	}

	if req.FirstName != nil && *req.FirstName != "" {
		input.UserAttributes = append(input.UserAttributes, types.AttributeType{
			Name:  aws.String("given_name"),
			Value: req.FirstName,
		})
	}

	if req.LastName != nil && *req.LastName != "" {
		input.UserAttributes = append(input.UserAttributes, types.AttributeType{
			Name:  aws.String("family_name"),
			Value: req.LastName,
		})
	}

	result, err := c.client.SignUp(ctx, input)
	if err != nil {
		return nil, c.mapCognitoError(err)
	}

	// For signup, we might need confirmation, so we don't have tokens yet
	user := &User{
		ID:       aws.ToString(result.UserSub),
		Email:    req.Email,
		Username: req.Email,
		Status:   UserStatusPending,
	}

	if req.FirstName != nil {
		user.FirstName = req.FirstName
	}
	if req.LastName != nil {
		user.LastName = req.LastName
	}

	return &AuthResult{
		User: user,
	}, nil
}

// SignIn authenticates a user with Cognito
func (c *CognitoProvider) SignIn(ctx context.Context, req *SignInRequest) (*AuthResult, error) {
	input := &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: types.AuthFlowTypeUserPasswordAuth,
		ClientId: aws.String(c.config.ClientID),
		AuthParameters: map[string]string{
			"USERNAME": req.Username,
			"PASSWORD": req.Password,
		},
	}

	result, err := c.client.InitiateAuth(ctx, input)
	if err != nil {
		return nil, c.mapCognitoError(err)
	}

	if result.AuthenticationResult == nil {
		return nil, errors.New("authentication failed")
	}

	// Get user details - assuming username is email for Cognito
	user, err := c.GetUserByEmail(ctx, req.Username)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		AccessToken:  aws.ToString(result.AuthenticationResult.AccessToken),
		RefreshToken: aws.ToString(result.AuthenticationResult.RefreshToken),
		IDToken:      aws.ToString(result.AuthenticationResult.IdToken),
		ExpiresIn:    int(result.AuthenticationResult.ExpiresIn),
		TokenType:    aws.ToString(result.AuthenticationResult.TokenType),
		User:         user,
	}, nil
}

// RefreshToken refreshes an access token using a refresh token
func (c *CognitoProvider) RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error) {
	input := &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: types.AuthFlowTypeRefreshTokenAuth,
		ClientId: aws.String(c.config.ClientID),
		AuthParameters: map[string]string{
			"REFRESH_TOKEN": refreshToken,
		},
	}

	result, err := c.client.InitiateAuth(ctx, input)
	if err != nil {
		return nil, c.mapCognitoError(err)
	}

	if result.AuthenticationResult == nil {
		return nil, errors.New("token refresh failed")
	}

	return &AuthResult{
		AccessToken:  aws.ToString(result.AuthenticationResult.AccessToken),
		RefreshToken: refreshToken,
		IDToken:      aws.ToString(result.AuthenticationResult.IdToken),
		ExpiresIn:    int(result.AuthenticationResult.ExpiresIn),
		TokenType:    aws.ToString(result.AuthenticationResult.TokenType),
	}, nil
}

// SignOut signs out a user from Cognito
func (c *CognitoProvider) SignOut(ctx context.Context, accessToken string) error {
	input := &cognitoidentityprovider.GlobalSignOutInput{
		AccessToken: aws.String(accessToken),
	}

	_, err := c.client.GlobalSignOut(ctx, input)
	return c.mapCognitoError(err)
}

// ValidateToken validates a JWT token with Cognito
func (c *CognitoProvider) ValidateToken(ctx context.Context, token string) (*TokenClaims, error) {
	input := &cognitoidentityprovider.GetUserInput{
		AccessToken: aws.String(token),
	}

	result, err := c.client.GetUser(ctx, input)
	if err != nil {
		return nil, c.mapCognitoError(err)
	}

	claims := &TokenClaims{
		UserID:    aws.ToString(result.Username),
		Username:  aws.ToString(result.Username),
		ExpiresAt: time.Now().Add(time.Hour), // Should parse from actual token
		IssuedAt:  time.Now(),
	}

	// Extract email from attributes
	for _, attr := range result.UserAttributes {
		if aws.ToString(attr.Name) == "email" {
			claims.Email = aws.ToString(attr.Value)
			break
		}
	}

	return claims, nil
}

// ResetPassword initiates password reset for a user
func (c *CognitoProvider) ResetPassword(ctx context.Context, email string) error {
	input := &cognitoidentityprovider.ForgotPasswordInput{
		ClientId: aws.String(c.config.ClientID),
		Username: aws.String(email),
	}

	_, err := c.client.ForgotPassword(ctx, input)
	return c.mapCognitoError(err)
}

// ConfirmPasswordReset confirms a password reset with a code
func (c *CognitoProvider) ConfirmPasswordReset(ctx context.Context, req *PasswordResetRequest) error {
	input := &cognitoidentityprovider.ConfirmForgotPasswordInput{
		ClientId:         aws.String(c.config.ClientID),
		Username:         aws.String(req.Email),
		ConfirmationCode: aws.String(req.ConfirmationCode),
		Password:         aws.String(req.NewPassword),
	}

	_, err := c.client.ConfirmForgotPassword(ctx, input)
	return c.mapCognitoError(err)
}

// Helper methods

// mapCognitoUserToUser converts a Cognito user to our User model
func (c *CognitoProvider) mapCognitoUserToUser(cognitoUser *types.UserType) *User {
	user := &User{
		ID:        aws.ToString(cognitoUser.Username),
		Username:  aws.ToString(cognitoUser.Username),
		Status:    c.mapCognitoUserStatus(cognitoUser.UserStatus),
		CreatedAt: aws.ToTime(cognitoUser.UserCreateDate),
		UpdatedAt: aws.ToTime(cognitoUser.UserLastModifiedDate),
	}

	// Extract attributes
	for _, attr := range cognitoUser.Attributes {
		switch aws.ToString(attr.Name) {
		case "email":
			user.Email = aws.ToString(attr.Value)
		case "given_name":
			user.FirstName = attr.Value
		case "family_name":
			user.LastName = attr.Value
		}
	}

	return user
}

// mapCognitoError maps Cognito errors to our standard errors
func (c *CognitoProvider) mapCognitoError(err error) error {
	if err == nil {
		return nil
	}

	var cognitoErr *types.UserNotFoundException
	if errors.As(err, &cognitoErr) {
		return ErrUserNotFound
	}

	var authErr *types.NotAuthorizedException
	if errors.As(err, &authErr) {
		return ErrInvalidCredentials
	}

	var existsErr *types.UsernameExistsException
	if errors.As(err, &existsErr) {
		return ErrUserAlreadyExists
	}

	return err
}

// mapCognitoUserStatus maps Cognito user status to our UserStatus
func (c *CognitoProvider) mapCognitoUserStatus(status types.UserStatusType) UserStatus {
	switch status {
	case types.UserStatusTypeConfirmed:
		return UserStatusActive
	case types.UserStatusTypeUnconfirmed:
		return UserStatusPending
	case types.UserStatusTypeArchived:
		return UserStatusInactive
	case types.UserStatusTypeCompromised:
		return UserStatusSuspended
	case types.UserStatusTypeUnknown:
		return UserStatusPending
	case types.UserStatusTypeResetRequired:
		return UserStatusPending
	case types.UserStatusTypeForceChangePassword:
		return UserStatusPending
	default:
		return UserStatusPending
	}
}

// Define standard errors
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
)
