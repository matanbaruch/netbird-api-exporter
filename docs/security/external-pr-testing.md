# External PR Testing Security Guide

## Overview

This guide explains how to securely handle pull requests from external contributors that require access to repository secrets for testing.

## The Problem

External contributors cannot access repository secrets (like `NETBIRD_API_TOKEN`) for security reasons. This prevents integration tests from running on their PRs, which could miss important issues.

## Our Solution

We've implemented a secure approval system that:

1. **Runs basic tests automatically** - Unit tests, linting, Docker builds (no secrets needed)
2. **Requires manual approval** - Integration tests with secrets only run after maintainer approval
3. **Uses label-based approval** - Simple workflow using GitHub labels
4. **Provides clear feedback** - Contributors know what's happening and why

## Security Features

### 🛡️ Security Safeguards

- **Manual Review Required**: All external PRs require code review before secrets access
- **Label-Based Approval**: Uses `approved-for-testing` label to control access
- **Environment Protection**: Uses GitHub environments for additional security
- **Audit Trail**: All approvals are logged with reviewer information
- **Clear Documentation**: Contributors understand the process

### 🔍 What Gets Tested

**Without Approval (Automatic):**

- ✅ Unit tests
- ✅ Linting and code quality
- ✅ Docker image builds
- ✅ Helm chart validation (with dummy values)
- ✅ Performance and benchmark tests

**With Approval (Comment `/approve` or add label):**

- 🔒 Integration tests with real NetBird API
- 🔒 Docker Compose full stack tests
- 🔒 Helm chart tests with real secrets
- 🔒 End-to-end functionality tests

## For Maintainers

### Quick Approval Process

1. **Review the PR code changes thoroughly**
2. **Choose your preferred approval method:**

**Option 1: Comment Approval (Recommended)**

```
/approve
```

Or comment any of these:

- `/approve-testing`
- `approved for testing`
- `approve for testing`

**Option 2: Manual Label**
Add the `approved-for-testing` label via GitHub UI

3. **Monitor test results**

### Security Checklist

Before approving any external PR, ensure:

- [ ] **Code Review Complete**: All changes have been reviewed
- [ ] **No Malicious Code**: No suspicious or harmful code patterns
- [ ] **No Secret Exposure**: Code doesn't log or expose secrets
- [ ] **Trusted Contributor**: Author has good GitHub history
- [ ] **Reasonable Changes**: Changes are within scope and reasonable
- [ ] **No Workflow Modifications**: No changes to `.github/workflows/` without extra scrutiny

## For Contributors

### What to Expect

When you open a PR from a fork:

1. **Basic tests run immediately** - No waiting needed
2. **Security check comment appears** - Explains the process
3. **Integration tests require approval** - Need maintainer review
4. **Maintainer can approve via comment** - Simple `/approve` comment
5. **Clear status updates** - You'll see what's happening

### Speeding Up Approval

To get faster approval:

- **Write clear PR descriptions** - Explain what and why
- **Keep changes focused** - Smaller PRs are easier to review
- **Add tests** - Show you've tested your changes
- **Follow guidelines** - Check [CONTRIBUTING.md](../../CONTRIBUTING.md)
- **Be patient** - Maintainers review as time permits

### After Approval

Once approved:

- **All tests run automatically** - Including integration tests
- **Future commits auto-approved** - No need to re-approve
- **Full test feedback** - You'll see all test results

## Technical Implementation

### Workflow Files

- **`.github/workflows/external-pr.yml`** - Main external PR workflow
- **Security first approach** - Uses `pull_request_target` safely
- **Environment protection** - Uses GitHub environments
- **Comprehensive testing** - Covers all test scenarios

### Security Model

```
External PR → Security Check → Basic Tests (Always)
                          ↓
                     Approval Required → Full Tests (If Approved)
                          ↓
                     Test Results → Feedback to PR
```

### Label Management

- **`approved-for-testing`** - Main approval label
- **Persistent approval** - Applies to all future commits
- **Easy to manage** - Can be added/removed as needed

## Monitoring and Auditing

### Test Results

All test results are:

- **Visible in PR** - Comments and checks show status
- **Documented in summaries** - Clear pass/fail information
- **Linked to workflows** - Easy to debug failures

### Approval Audit

- **GitHub audit log** - All label changes are logged
- **PR comments** - Optional approval comments with timestamps
- **Workflow history** - Complete test execution history

## Troubleshooting

### Common Issues

**Can't add label:**

- Check repository permissions
- Ensure you're a maintainer
- Try adding label manually in GitHub UI

**Tests still don't run:**

- Verify label was added successfully
- Check workflow run logs
- Ensure PR is from external repository

### Getting Help

- **Check workflow logs** - Look for error messages
- **Review PR comments** - Automated feedback explains issues
- **Ask in discussions** - Use GitHub Discussions for questions

## Best Practices

### For Maintainers

1. **Regular Review Schedule** - Check for external PRs regularly
2. **Document Decisions** - Use approval comments for audit trail
3. **Monitor Test Results** - Follow up on test failures
4. **Update Documentation** - Keep this guide current

### For Contributors

1. **Follow Contributing Guidelines** - Read [CONTRIBUTING.md](../../CONTRIBUTING.md)
2. **Test Locally First** - Run tests before submitting
3. **Clear Communication** - Explain changes in PR description
4. **Be Responsive** - Address review feedback promptly

## Security Considerations

### What This Protects Against

- **Credential Theft** - Prevents malicious code from accessing secrets
- **Resource Abuse** - Limits who can run expensive tests
- **Supply Chain Attacks** - Requires human review of all external code
- **Accidental Exposure** - Prevents secrets from being logged or exposed

### What This Doesn't Protect Against

- **Social Engineering** - Still requires careful code review
- **Compromised Maintainer** - Depends on maintainer security
- **Repository Compromise** - Broader security measures needed

### Additional Security Measures

Consider implementing:

- **Two-person approval** - Require multiple maintainer approvals
- **Time delays** - Wait period before approval takes effect
- **Automated security scanning** - Additional security checks
- **Limited scope secrets** - Use test-specific tokens when possible

---

## Summary

This system balances security with contributor experience by:

- **Keeping barriers low** - Most tests run without approval
- **Requiring human review** - Sensitive tests need approval
- **Providing clear feedback** - Everyone knows what's happening
- **Maintaining security** - Secrets are protected from malicious code

The result is a secure, maintainable process that welcomes external contributions while protecting repository secrets and resources.
