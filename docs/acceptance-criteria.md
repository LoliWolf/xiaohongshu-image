# Acceptance Criteria

This document outlines the acceptance criteria for the Xiaohongshu Image Generation System.

## 1. MockConnector Full Workflow

### Criteria
- [ ] System can "pull" new comments and store them in the database
- [ ] Comments table shows new entries after polling
- [ ] Polling can be triggered manually via API
- [ ] Polling runs automatically on schedule

### Verification Steps

1. Start all services:
   ```bash
   docker-compose up -d
   ```

2. Access settings page: http://localhost:3000/settings
   - Confirm Connector Mode is "mock"

3. Click "Run Poll Now" button

4. Check database:
   ```bash
   docker-compose exec mysql mysql -u root -prootpassword xiaohongshu_image -e "SELECT * FROM comments;"
   ```

5. Verify comments are stored with correct data:
   - comment_uid
   - user_name
   - content
   - ingested_at

**Expected Result**: 6 mock comments should be stored in the database.

---

## 2. Comment with Email + Clear Keywords

### Criteria
- [ ] IntentJob passes and creates a task
- [ ] Task status is EXTRACTED
- [ ] Email is extracted correctly
- [ ] Prompt is extracted correctly
- [ ] Confidence score is >= threshold

### Verification Steps

1. Ensure polling has run (see AC #1)

2. Access tasks page: http://localhost:3000/tasks

3. Verify tasks are created for comments with:
   - Email addresses
   - Generation keywords (出图, 生成图, 做视频, etc.)

4. Click on a task to view details

5. Verify task details:
   - Status: EXTRACTED
   - Email: valid email address
   - Prompt: extracted from comment
   - Confidence: >= 0.7
   - Request Type: image or video

**Expected Result**: 4 tasks should be created (comments 1, 2, 4, 5, 6).

---

## 3. MockProvider Generation

### Criteria
- [ ] Task status changes to SUCCEEDED within 10-30 seconds
- [ ] result_url is populated
- [ ] result_url is accessible (MinIO signed link or /api/files)
- [ ] Provider job ID is stored

### Verification Steps

1. Monitor a task on the tasks page (auto-refresh enabled)

2. Wait for status to change from EXTRACTED → SUBMITTED → RUNNING → SUCCEEDED

3. Click on the task to view details

4. Verify:
   - Status: SUCCEEDED
   - Provider Name: mock
   - Provider Job ID: mock_job_*
   - Result URL: http://localhost:9000/... (MinIO URL)

5. Click the Result URL to verify it's accessible

**Expected Result**: Task completes within 10-30 seconds with valid result URL.

---

## 4. Email Delivery

### Criteria
- [ ] Mailhog/SMTP receives email with result_url
- [ ] Task status changes to EMAILED
- [ ] Deliveries table records SENT status
- [ ] Email content includes:
  - Request type (图片/视频)
  - Prompt
  - Download link
  - Expiry notice

### Verification Steps

1. Wait for a task to reach SUCCEEDED status

2. Access Mailhog: http://localhost:8025

3. Verify email is received:
   - Check inbox
   - Verify sender: noreply@xiaohongshu-image.local
   - Verify recipient matches comment email

4. Open email and verify content:
   - Subject: 您的图片生成结果已就绪 / 您的视频生成结果已就绪
   - Body includes prompt
   - Body includes download link
   - Body mentions 1-hour expiry

5. Check task details page:
   - Status: EMAILED
   - Deliveries section shows SENT status
   - Sent at timestamp is populated

6. Check database:
   ```bash
   docker-compose exec mysql mysql -u root -prootpassword xiaohongshu_image -e "SELECT * FROM deliveries;"
   ```

**Expected Result**: Email is received in Mailhog with correct content, task status is EMAILED.

---

## 5. Idempotency

### Criteria
- [ ] Same comment_uid doesn't create duplicate tasks
- [ ] Same comment_uid doesn't send duplicate emails
- [ ] Database constraints prevent duplicates

### Verification Steps

1. Click "Run Poll Now" multiple times

2. Check comments table:
   ```bash
   docker-compose exec mysql mysql -u root -prootpassword xiaohongshu_image -e "SELECT COUNT(*) FROM comments;"
   ```

3. Check tasks table:
   ```bash
   docker-compose exec mysql mysql -u root -prootpassword xiaohongshu_image -e "SELECT COUNT(*) FROM tasks;"
   ```

4. Check deliveries table:
   ```bash
   docker-compose exec mysql mysql -u root -prootpassword xiaohongshu_image -e "SELECT COUNT(*) FROM deliveries;"
   ```

5. Verify counts don't increase after repeated polling

**Expected Result**: Counts remain the same after multiple poll runs.

---

## 6. Error Visibility

### Criteria
- [ ] LLM/Provider failures set task status to FAILED
- [ ] Error message is recorded
- [ ] Queue doesn't retry infinitely
- [ ] Error is visible in task details
- [ ] Audit log records error events

### Verification Steps

#### Scenario 1: LLM Failure

1. Set invalid LLM API key in settings
2. Trigger poll
3. Check task status
4. Verify error message is visible

#### Scenario 2: Provider Failure

1. Configure invalid Provider URL
2. Wait for task processing
3. Check task status
4. Verify error message is visible

#### Scenario 3: Max Retries Exceeded

1. Monitor a task that keeps failing
2. Wait for 20 status checks
3. Verify task status becomes FAILED
4. Verify error: "max retries exceeded"

**Expected Result**: All failures are properly recorded and visible in UI.

---

## 7. Configuration Management

### Criteria
- [ ] Settings can be retrieved via API
- [ ] Settings can be updated via API
- [ ] Changes persist across restarts
- [ ] UI reflects current settings

### Verification Steps

1. Access settings page: http://localhost:3000/settings

2. Modify a setting (e.g., polling_interval_sec)

3. Click "Save Settings"

4. Refresh page and verify change persists

5. Restart services:
   ```bash
   docker-compose restart api worker
   ```

6. Access settings page again

7. Verify settings are still updated

**Expected Result**: Settings persist correctly.

---

## 8. Real MCP Connector (Optional)

### Criteria
- [ ] Can switch to MCP mode
- [ ] Can configure MCP server URL
- [ ] System connects to MCP server
- [ ] Comments are fetched from real source

### Verification Steps

1. Access settings page

2. Change Connector Mode to "mcp"

3. Enter MCP server URL

4. Save settings

5. Trigger poll

6. Check logs for MCP connection

**Expected Result**: System connects to MCP server and fetches real comments.

---

## 9. Provider Mapping

### Criteria
- [ ] Can configure custom provider
- [ ] Request mapping works correctly
- [ ] Response mapping works correctly
- [ ] Status mapping works correctly

### Verification Steps

1. Access settings page

2. Configure a custom provider in Provider JSON:

```json
[
  {
    "provider_name": "test-provider",
    "type": "both",
    "base_url": "https://httpbin.org",
    "api_key": "",
    "submit_path": "/post",
    "status_path_template": "/anything/{id}",
    "headers": {},
    "request_mapping": {
      "data": {
        "prompt_text": "$.prompt",
        "type": "$.type"
      }
    },
    "response_mapping": {
      "job_id_jsonpath": "$.json.data"
    },
    "status_mapping": {
      "status_jsonpath": "$.status",
      "result_url_jsonpath": "$.url"
    }
  }
]
```

3. Save settings

4. Trigger poll to create a task

5. Monitor task processing

**Expected Result**: Provider mapping transforms request/response correctly.

---

## 10. Performance and Scalability

### Criteria
- [ ] System handles 100+ concurrent comments
- [ ] Worker processes jobs in parallel
- [ ] Database queries are efficient
- [ ] No memory leaks

### Verification Steps

1. Generate 100 mock comments

2. Trigger poll

3. Monitor system resources:
   ```bash
   docker stats
   ```

4. Check task processing time

5. Verify all tasks complete

**Expected Result**: System handles load without performance degradation.

---

## Summary Checklist

### Core Functionality
- [ ] MockConnector workflow works end-to-end
- [ ] Intent recognition works correctly
- [ ] Task creation and processing works
- [ ] Email delivery works
- [ ] Idempotency is maintained

### Error Handling
- [ ] Errors are visible in UI
- [ ] Errors are logged properly
- [ ] Retries work as expected
- [ ] Max retry limit is enforced

### Configuration
- [ ] Settings can be updated
- [ ] Settings persist correctly
- [ ] Provider mapping works
- [ ] MCP connector can be configured

### Performance
- [ ] System handles expected load
- [ ] Response times are acceptable
- [ ] Resources are used efficiently
- [ ] No memory leaks

---

## Sign-off

**Tester**: _________________ **Date**: _________________

**Reviewer**: _________________ **Date**: _________________

**Approved**: ☐ Yes  ☐ No  ☐ With Changes
