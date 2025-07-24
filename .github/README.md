# GitHub Actions CI/CD Setup Complete ✅

Your infrastructure has been successfully deployed with complete GitHub Actions integration. The CI/CD pipeline is now ready to use with the following setup:

## 🎯 Pipeline Behavior

**Main Branch (`main`):**
- ✅ Lint → ✅ Test → ✅ Build
- ✅ Build & Push to ECR (parallel)
- ✅ Deploy Infrastructure (parallel)

**Feature Branches:**
- ✅ Lint → ✅ Test → ✅ Build (no ECR push or infrastructure deployment)

## 🔑 Required GitHub Secrets

You need to configure these secrets in your GitHub repository:

### Step 1: Go to GitHub Repository Settings
1. Navigate to your GitHub repository
2. Go to **Settings** → **Secrets and variables** → **Actions**
3. Click **New repository secret**

### Step 2: Add Required Secrets
Add these three secrets:

| Secret Name | Purpose | Value |
|-------------|---------|-------|
| `AWS_ROLE_ARN` | ECR operations only | `arn:aws:iam::698315877107:role/CodeRefactor-GitHubActions-ECR-Role` |
| `AWS_INFRA_ROLE_ARN` | Infrastructure deployment | `arn:aws:iam::698315877107:role/CodeRefactor-GitHubActions-Infrastructure-Role` |
| `AWS_REGION` | AWS region | `us-east-1` |

## 🏗️ Initial Deployment Strategy

To avoid the chicken-and-egg problem with ECS Fargate and ECR, the infrastructure is deployed with the following approach:

1. **Infrastructure Deployment**: Creates all resources including ECR repository and ECS service with **0 desired tasks**
2. **First GitHub Actions Run**: Builds and pushes your actual application image to ECR  
3. **ECS Service Scale-Up**: After successful image push, manually scale the ECS service to 1 task or use GitHub Actions to update the service

### Manual Scale-Up After First Deployment:
```bash
aws ecs update-service --cluster <cluster-name> --service CodeRefactorService --desired-count 1
```

## 🚀 How It Works

1. **GitHub Actions OIDC**: Your infrastructure includes a GitHub Actions OIDC provider that allows secure authentication without long-lived credentials
2. **Dual IAM Roles**: 
   - **ECR Role**: Limited permissions for Docker image operations
   - **Infrastructure Role**: Full permissions for CDK/CloudFormation deployments
3. **Parallel Execution**: On main branch, ECR push and infrastructure deployment run simultaneously for faster deployments
4. **Conditional Pipeline**: The workflow automatically detects if you're on `main` branch and runs deployment jobs only then

## 🧪 Testing the Pipeline

1. **Feature Branch Test**: Create a new branch, make changes, and push - should run lint/test/build only
2. **Main Branch Test**: Merge to main - should run lint/test/build, then ECR push + infrastructure deployment in parallel

## 📋 Infrastructure Outputs

Your deployed infrastructure provides these resources:
- **ECR Repository**: For Docker images
- **GitHub Actions Roles**: For secure CI/CD access (ECR + Infrastructure)
- **OIDC Provider**: For passwordless authentication
- **Complete Application Stack**: Including RDS, S3, ECS, API Gateway, etc.

## ✨ Next Steps

Once you set the GitHub secrets, your CI/CD pipeline will be fully operational. Every push to main will automatically:
1. Build and deploy your Docker image to ECR
2. Deploy any infrastructure changes via CDK

Both operations run in parallel for faster deployments!

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::YOUR_ACCOUNT_ID:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
        },
        "StringLike": {
          "token.actions.githubusercontent.com:sub": "repo:YOUR_GITHUB_ORG/YOUR_REPO_NAME:ref:refs/heads/main"
        }
      }
    }
  ]
}
```

2. Attach the following policy to the role:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecr:GetAuthorizationToken",
        "ecr:BatchCheckLayerAvailability",
        "ecr:GetDownloadUrlForLayer",
        "ecr:BatchGetImage",
        "ecr:PutImage",
        "ecr:InitiateLayerUpload",
        "ecr:UploadLayerPart",
        "ecr:CompleteLayerUpload"
      ],
      "Resource": "*"
    }
  ]
}
```

### Setting up GitHub OIDC Provider (if not already done)

If you haven't set up the OIDC provider in your AWS account, run this AWS CLI command:

```bash
aws iam create-open-id-connect-provider \
  --url https://token.actions.githubusercontent.com \
  --client-id-list sts.amazonaws.com \
  --thumbprint-list 6938fd4d98bab03faadb97b34396831e3780aea1
```

## Summary

✅ **Infrastructure Complete**: Both the OIDC provider and GitHub Actions IAM role are now managed by your CDK stack in `app_stack.go`.

### GitHub Repository Secrets (Required)

Set these secrets in your GitHub repository at `Settings` → `Secrets and variables` → `Actions`:

1. **AWS_ROLE_ARN**: `arn:aws:iam::698315877107:role/CodeRefactorInfra-GitHubActionsECRRole9AA9BEF6-iiIFITW4L575`
2. **AWS_REGION**: `us-east-1`

### Workflow Behavior

- **Main Branch Push**: Runs lint → test → build → push to ECR
- **Other Branches/PRs**: Runs lint → test → build (no ECR push)

The ECR repository name is configured as `refactor-ecr-repo` which matches your CDK infrastructure.

---

## Manual Setup (No Longer Needed)

The sections below are kept for reference, but you don't need to do this manually since everything is now in your CDK stack.
