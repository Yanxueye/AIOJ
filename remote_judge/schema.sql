CREATE TABLE IF NOT EXISTS problems (
    id BIGINT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    time_limit_ms INT NOT NULL,
    memory_limit_mb INT NOT NULL,
    output_limit_kb INT NOT NULL
);

CREATE TABLE IF NOT EXISTS test_cases (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    problem_id BIGINT NOT NULL,
    case_no INT NOT NULL,
    input_text MEDIUMTEXT NOT NULL,
    expected_text MEDIUMTEXT NOT NULL,
    UNIQUE KEY uk_problem_case (problem_id, case_no),
    CONSTRAINT fk_test_cases_problem FOREIGN KEY (problem_id) REFERENCES problems(id)
);

CREATE TABLE IF NOT EXISTS submissions (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    problem_id BIGINT NOT NULL,
    trace_id VARCHAR(128) NOT NULL,
    language VARCHAR(64) NOT NULL,
    code MEDIUMTEXT NOT NULL,
    code_length INT NOT NULL,
    status VARCHAR(64) NOT NULL,
    runtime_ms INT NOT NULL DEFAULT 0,
    memory_kb INT NOT NULL DEFAULT 0,
    compile_output MEDIUMTEXT NULL,
    error_message VARCHAR(512) NULL,
    queue_started_at DATETIME NULL,
    judge_started_at DATETIME NULL,
    finished_at DATETIME NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    INDEX idx_submissions_user_created (user_id, created_at),
    INDEX idx_submissions_problem (problem_id),
    INDEX idx_submissions_status (status),
    INDEX idx_submissions_trace (trace_id)
);

CREATE TABLE IF NOT EXISTS submission_case_results (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    submission_id BIGINT NOT NULL,
    case_no INT NOT NULL,
    status VARCHAR(64) NOT NULL,
    runtime_ms INT NOT NULL DEFAULT 0,
    memory_kb INT NOT NULL DEFAULT 0,
    stdout_bytes INT NOT NULL DEFAULT 0,
    stderr_bytes INT NOT NULL DEFAULT 0,
    signal_name VARCHAR(64) NULL,
    stdout_preview MEDIUMTEXT NULL,
    stderr_preview MEDIUMTEXT NULL,
    UNIQUE KEY uk_submission_case (submission_id, case_no),
    CONSTRAINT fk_case_results_submission FOREIGN KEY (submission_id) REFERENCES submissions(id)
);

INSERT INTO problems (id, title, time_limit_ms, memory_limit_mb, output_limit_kb)
VALUES
    (1001, 'A+B Problem', 1000, 128, 1024),
    (1002, 'Echo', 1000, 128, 1024)
ON DUPLICATE KEY UPDATE
    title = VALUES(title),
    time_limit_ms = VALUES(time_limit_ms),
    memory_limit_mb = VALUES(memory_limit_mb),
    output_limit_kb = VALUES(output_limit_kb);

INSERT INTO test_cases (problem_id, case_no, input_text, expected_text)
VALUES
    (1001, 1, '1 2\n', '3\n'),
    (1001, 2, '10 20\n', '30\n'),
    (1002, 1, 'hello\n', 'hello\n')
ON DUPLICATE KEY UPDATE
    input_text = VALUES(input_text),
    expected_text = VALUES(expected_text);
