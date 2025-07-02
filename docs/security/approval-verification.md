# External PR Approval System Verification

This document verifies that the external PR approval system properly restricts access to secrets and only allows maintainers to approve PRs for testing.

## 🔐 Security Verification Checklist

### ✅ Permission Checking

**Requirement:** Only maintainers with `write`, `maintain`, or `admin` permissions can approve external PRs.

**Implementation:**

- Uses GitHub API `getCollaboratorPermissionLevel` to check user permissions
- Only allows approval if permission is one of: `['admin', 'maintain', 'write']`
- Provides clear error messages for unauthorized users

**Code Location:** `.github/workflows/external-pr.yml` - `comment-approval` job

### ✅ External PR Detection

**Requirement:** Only external PRs (from forks) should require approval.

**Implementation:**

- Compares `pr.head.repo.full_name` with current repository
- Internal PRs automatically approved
- External PRs require manual approval

**Code Location:** `.github/workflows/external-pr.yml` - `security-check` job

### ✅ Comment-Based Approval

**Requirement:** Maintainers can approve via comments with specific keywords.

**Implementation:**

- Monitors for approval keywords: `/approve`, `/approve-testing`, `approved for testing`, etc.
- Validates commenter permissions before processing
- Adds `approved-for-testing` label automatically
- Posts confirmation comment with audit trail

**Supported Commands:**

- `/approve` (recommended)
- `/approve-testing`
- `approved for testing`
- `approve for testing`
- `/test-approved`

### ✅ Clear Instructions

**Requirement:** Maintainers receive clear instructions on how to approve PRs.

**Implementation:**

- Automatic comment posted on external PRs
- Step-by-step approval instructions
- Security checklist for reviewers
- Clear distinction between approved and unapproved tests

## 🧪 Test Scenarios

### Scenario 1: External PR from Fork

**Expected Behavior:**

1. PR opened from fork triggers security check
2. Automatic comment posted with approval instructions
3. Basic tests run immediately (no secrets)
4. Integration tests wait for approval

**Verification Steps:**

1. Create PR from fork
2. Verify security comment appears
3. Verify basic tests run
4. Verify integration tests don't run until approved

### Scenario 2: Maintainer Approval via Comment

**Expected Behavior:**

1. Maintainer comments `/approve`
2. System checks maintainer permissions
3. Label added automatically
4. Confirmation comment posted
5. Full test suite runs

**Verification Steps:**

1. Maintainer comments `/approve` on external PR
2. Verify permission check passes
3. Verify `approved-for-testing` label added
4. Verify confirmation comment posted
5. Verify full tests run

### Scenario 3: Non-Maintainer Attempts Approval

**Expected Behavior:**

1. Non-maintainer comments `/approve`
2. System checks permissions
3. Permission denied message posted
4. No approval granted
5. Tests remain restricted

**Verification Steps:**

1. External contributor or read-only user comments `/approve`
2. Verify permission check fails
3. Verify error message explains permission requirements
4. Verify no label added
5. Verify tests remain restricted

### Scenario 4: Internal PR (Same Repository)

**Expected Behavior:**

1. PR opened from same repository
2. No approval required
3. All tests run automatically
4. No security comment posted

**Verification Steps:**

1. Create PR from branch in same repository
2. Verify no approval workflow triggered
3. Verify all tests run immediately
4. Verify regular test workflows handle the PR

## 🛡️ Security Controls Summary

| Control                   | Implementation                              | Status         |
| ------------------------- | ------------------------------------------- | -------------- |
| **Permission Validation** | GitHub API `getCollaboratorPermissionLevel` | ✅ Implemented |
| **External PR Detection** | Repository comparison logic                 | ✅ Implemented |
| **Approval Keywords**     | Pattern matching with validation            | ✅ Implemented |
| **Audit Trail**           | Approval comments with timestamps           | ✅ Implemented |
| **Label Management**      | Automatic label application                 | ✅ Implemented |
| **Clear Instructions**    | Enhanced comment formatting                 | ✅ Implemented |
| **Error Handling**        | Permission denied messages                  | ✅ Implemented |
| **Secret Protection**     | Environment-based access control            | ✅ Implemented |

## 📋 Manual Testing Guide

### For Repository Maintainers

1. **Test External PR Flow:**

   ```bash
   # 1. Have external contributor create PR from fork
   # 2. Verify security comment appears
   # 3. Comment "/approve" on the PR
   # 4. Verify full test suite runs
   ```

2. **Test Permission Boundaries:**

   ```bash
   # 1. Ask read-only user to comment "/approve"
   # 2. Verify permission denied message
   # 3. Verify no tests run with secrets
   ```

3. **Test Different Approval Methods:**
   ```bash
   # Test various approval commands:
   # - "/approve"
   # - "/approve-testing"
   # - "approved for testing"
   # - Manual label addition
   ```

### For External Contributors

1. **Understand the Process:**

   - Open PR from fork
   - Read security comment explanation
   - Wait for maintainer approval
   - See basic tests run immediately
   - Understand which tests need approval

2. **Verify Basic Tests Run:**
   - Unit tests should run without approval
   - Linting should run without approval
   - Docker build should run without approval
   - Integration tests should wait for approval

## 🔍 Monitoring and Alerts

### Key Metrics to Monitor

- **Approval Rate:** How many external PRs get approved
- **Time to Approval:** How long maintainers take to approve
- **Permission Errors:** Failed approval attempts
- **Test Coverage:** Success rate of approved PRs

### Audit Points

- All approval actions logged in PR comments
- GitHub audit log captures label changes
- Workflow run history shows test execution
- Permission checks logged in workflow outputs

## 🚨 Security Incident Response

### If Unauthorized Approval Detected

1. **Immediate Actions:**

   - Remove `approved-for-testing` label
   - Cancel running workflows
   - Review PR changes thoroughly

2. **Investigation:**

   - Check workflow logs for permission validation
   - Verify user permissions in GitHub
   - Review audit log for suspicious activity

3. **Prevention:**
   - Review and update permission requirements
   - Consider additional approval requirements
   - Update security documentation

### If Secrets Potentially Exposed

1. **Immediate Actions:**

   - Rotate affected secrets immediately
   - Cancel all running workflows
   - Review logs for secret exposure

2. **Assessment:**

   - Determine scope of potential exposure
   - Check if secrets were logged or transmitted
   - Evaluate impact on systems

3. **Recovery:**
   - Update all affected secrets
   - Notify relevant stakeholders
   - Document incident and lessons learned

---

## ✅ Verification Status

**Last Verified:** $(date -u '+%Y-%m-%d %H:%M:%S UTC')
**Verification Method:** Code review and system design analysis
**Next Review:** Recommended after any changes to approval workflow

**Overall Security Rating:** ✅ **SECURE**

The external PR approval system implements proper security controls and follows GitHub security best practices for protecting repository secrets while maintaining contributor accessibility.
