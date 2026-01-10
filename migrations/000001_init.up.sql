CREATE TABLE IF NOT EXISTS settings (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    connector_mode VARCHAR(20) NOT NULL DEFAULT 'mock',
    mcp_server_cmd TEXT,
    mcp_server_url VARCHAR(500),
    mcp_auth TEXT,
    note_target VARCHAR(500) NOT NULL,
    polling_interval_sec INT NOT NULL DEFAULT 120,
    llm_base_url VARCHAR(500),
    llm_api_key VARCHAR(200),
    llm_model VARCHAR(100),
    llm_timeout_sec INT DEFAULT 15,
    intent_threshold DECIMAL(3,2) NOT NULL DEFAULT 0.70,
    smtp_host VARCHAR(200),
    smtp_port INT,
    smtp_user VARCHAR(200),
    smtp_pass VARCHAR(200),
    smtp_from VARCHAR(200),
    provider_json JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_settings (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS notes (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    note_target VARCHAR(500) NOT NULL UNIQUE,
    last_cursor TEXT,
    last_polled_at TIMESTAMP NULL,
    last_error TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS comments (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    note_target VARCHAR(500) NOT NULL,
    comment_uid VARCHAR(100) NOT NULL,
    user_name VARCHAR(200),
    content TEXT,
    comment_created_at TIMESTAMP NULL,
    ingested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_comment_uid (comment_uid),
    KEY idx_note_target (note_target),
    KEY idx_ingested_at (ingested_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS tasks (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    comment_id BIGINT UNSIGNED NOT NULL,
    status ENUM('PENDING', 'EXTRACTED', 'SUBMITTED', 'RUNNING', 'SUCCEEDED', 'EMAILED', 'FAILED') NOT NULL DEFAULT 'PENDING',
    request_type ENUM('image', 'video') NOT NULL,
    email VARCHAR(200),
    prompt TEXT,
    confidence DECIMAL(3,2),
    provider_name VARCHAR(100),
    provider_job_id VARCHAR(200),
    result_object_key VARCHAR(500),
    result_url VARCHAR(1000),
    error TEXT,
    retry_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_comment_id (comment_id),
    KEY idx_status_created (status, created_at),
    KEY idx_email (email),
    FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS deliveries (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    task_id BIGINT UNSIGNED NOT NULL,
    email_to VARCHAR(200) NOT NULL,
    status ENUM('SENT', 'FAILED') NOT NULL,
    sent_at TIMESTAMP NULL,
    error TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    KEY idx_task_id (task_id),
    KEY idx_status (status),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    level VARCHAR(20) NOT NULL,
    event VARCHAR(100) NOT NULL,
    payload_json JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    KEY idx_event (event),
    KEY idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO settings (
    connector_mode,
    note_target,
    polling_interval_sec,
    llm_base_url,
    llm_model,
    llm_timeout_sec,
    intent_threshold,
    provider_json
) VALUES (
    'mock',
    'https://www.xiaohongshu.com/explore/12345678',
    120,
    'https://api.openai.com/v1',
    'gpt-4o-mini',
    15,
    0.70,
    JSON_ARRAY(
        JSON_OBJECT(
            'provider_name', 'mock',
            'type', 'both',
            'base_url', '',
            'api_key', '',
            'submit_path', '/jobs',
            'status_path_template', '/jobs/{id}',
            'headers', JSON_OBJECT(),
            'request_mapping', JSON_OBJECT(),
            'response_mapping', JSON_OBJECT('job_id_jsonpath', '$.data.id'),
            'status_mapping', JSON_OBJECT('status_jsonpath', '$.status', 'result_url_jsonpath', '$.output.url')
        )
    )
);
