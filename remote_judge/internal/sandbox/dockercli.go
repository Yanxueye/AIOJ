package sandbox

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// DockerCLISandbox 通过 Docker CLI 提供沙箱实现。
type DockerCLISandbox struct {
	TransferMode string
}

var imageReady sync.Map
var SeccompProfileMode = "embedded"

// Compile 在 Docker 容器中执行编译命令。
func (d *DockerCLISandbox) Compile(ctx context.Context, req ExecRequest) (ExecResult, error) {
	return d.exec(ctx, req)
}

// Run 在 Docker 容器中执行运行命令。
func (d *DockerCLISandbox) Run(ctx context.Context, req ExecRequest) (ExecResult, error) {
	return d.exec(ctx, req)
}

// Health 检查 Docker 守护进程是否可达。
func (d *DockerCLISandbox) Health(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "version", "--format", "{{.Server.Version}}")
	return cmd.Run()
}

func (d *DockerCLISandbox) exec(ctx context.Context, req ExecRequest) (ExecResult, error) {
	if len(req.Command) == 0 {
		return ExecResult{}, errors.New("empty command")
	}
	if err := d.ensureImage(ctx, req.Image); err != nil {
		return ExecResult{}, err
	}

	if d.transferMode() == "copy" {
		return d.execCopyMode(ctx, req)
	}
	return d.execBindMode(ctx, req)
}

func (d *DockerCLISandbox) execBindMode(ctx context.Context, req ExecRequest) (ExecResult, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, req.TimeLimit+2*time.Second)
	defer cancel()

	return d.dockerRun(timeoutCtx, req)
}

// dockerRun executes a command via `docker run --rm`, combining create+start+rm
// 相比 create/start/inspect/rm 分步调用节省约 30ms。
func (d *DockerCLISandbox) dockerRun(ctx context.Context, req ExecRequest) (ExecResult, error) {
	containerName := "judge_" + strconv.FormatInt(time.Now().UnixNano(), 36)
	seccompPath, cleanup, err := prepareSeccompProfile(ctx, SeccompProfileMode)
	if err != nil {
		return ExecResult{}, err
	}
	defer cleanup()

	// 从容器内部捕获 cgroup 内存峰值。包装器将读取
	// /sys/fs/cgroup/memory.peak 的命令追加到原始命令末尾，并将结果写入
	// /workspace 中的文件——该目录通过 bind-mount 挂载，宿主机可直接读取。
	memFile := ".judge_mem_" + containerName
	memHostPath := req.WorkDir + string(os.PathSeparator) + memFile
	wrappedCmd := wrapWithMemPeak(req.Command, "/workspace/"+memFile)

	args := []string{
		"run", "--rm",
		"--name", containerName,
		"-i",
		"--network", "none",
		"--cpus", "1",
		"--memory", strconv.Itoa(req.MemoryLimitMB) + "m",
		"--pids-limit", "64",
		"--tmpfs", "/tmp:size=64m",
		"--cap-drop", "ALL",
		"--security-opt", "no-new-privileges",
		"--read-only",
		"--user", "1000:1000",
		"-v", req.WorkDir + ":/workspace",
		"-w", "/workspace",
	}
	if seccompPath != "" {
		args = append(args, "--security-opt", "seccomp="+seccompPath)
	}
	args = append(args, req.Image)
	args = append(args, wrappedCmd...)

	start := time.Now()
	stdoutText, stderrText, exitCode, runErr := d.runDockerCmd(ctx, args, req.Stdin)

	// 读取容器内包装器写入的内存峰值文件。
	memoryKB := 0
	if data, err := os.ReadFile(memHostPath); err == nil {
		if peakBytes, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64); err == nil {
			memoryKB = int(peakBytes / 1024)
		}
		_ = os.Remove(memHostPath)
	}

	result := ExecResult{
		ExitCode:    exitCode,
		Stdout:      stdoutText,
		Stderr:      stderrText,
		Runtime:     time.Since(start),
		MemoryKB:    memoryKB,
		StdoutBytes: len(stdoutText),
		StderrBytes: len(stderrText),
		OOMKilled:   exitCode == 137,
	}
	if exitCode >= 128 {
		if sig := syscall.Signal(exitCode - 128); sig > 0 {
			result.Signal = sig.String()
		}
	}

	if ctx.Err() == context.DeadlineExceeded {
		result.TimedOut = true
		result.ExitCode = 124
		return result, nil
	}
	if runErr != nil && result.ExitCode == 0 {
		return result, runErr
	}
	return result, nil
}

func (d *DockerCLISandbox) runDockerCmd(ctx context.Context, args []string, stdin string) (string, string, int, error) {
	cmd := exec.CommandContext(ctx, "docker", args...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()

	exitCode := 0
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}
	return stdout.String(), stderr.String(), exitCode, runErr
}

func (d *DockerCLISandbox) execCopyMode(ctx context.Context, req ExecRequest) (ExecResult, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, req.TimeLimit+2*time.Second)
	defer cancel()

	containerID, err := d.createTaskContainer(timeoutCtx, req, []string{"sh", "-lc", "sleep 3600"})
	if err != nil {
		return ExecResult{}, err
	}
	defer d.removeContainer(context.Background(), containerID)

	if err := d.startDetached(timeoutCtx, containerID); err != nil {
		return ExecResult{}, err
	}
	if err := d.prepareWorkspace(timeoutCtx, containerID, req); err != nil {
		return ExecResult{}, err
	}

	baselineKB, _ := readCgroupMemoryPeak(timeoutCtx, containerID)

	start := time.Now()
	result, execErr := d.execInContainer(timeoutCtx, containerID, req.Command, req.Stdin)
	result.Runtime = time.Since(start)

	// 将命令产生的文件（如编译产物）回传到判题工作区，
	// 以便后续运行步骤使用。
	if result.ExitCode == 0 {
		_ = d.collectOutput(timeoutCtx, containerID, req.WorkDir)
	}

	finalKB, _ := readCgroupMemoryPeak(context.Background(), containerID)
	result.MemoryKB = finalKB - baselineKB
	if result.MemoryKB < 0 {
		result.MemoryKB = 0
	}

	inspect, inspectErr := d.inspectContainer(context.Background(), containerID)
	if inspectErr == nil {
		result.OOMKilled = inspect.State.OOMKilled
		if result.ExitCode == 0 {
			result.ExitCode = inspect.State.ExitCode
		}
		if inspect.State.Error != "" && result.Stderr == "" {
			result.Stderr = inspect.State.Error
			result.StderrBytes = len(result.Stderr)
		}
	}

	if timeoutCtx.Err() == context.DeadlineExceeded {
		result.TimedOut = true
		result.ExitCode = 124
		return result, nil
	}
	if execErr != nil && result.ExitCode == 0 {
		return result, execErr
	}
	return result, nil
}

// collectOutput 将沙箱容器 /workspace 中的文件复制回判题工作区。's /workspace back to
// 这是为了让编译产物在独立的沙箱容器（编译→运行）之间保留。

func (d *DockerCLISandbox) collectOutput(ctx context.Context, containerID, workDir string) error {
	listCmd := exec.CommandContext(ctx, "docker", "exec", containerID, "sh", "-lc", "ls -1 /workspace/")
	var listOut, listErr bytes.Buffer
	listCmd.Stdout = &listOut
	listCmd.Stderr = &listErr
	if err := listCmd.Run(); err != nil {
		return errors.New("collectOutput list: " + strings.TrimSpace(nonEmpty(listErr.String(), err.Error())))
	}
	names := strings.Fields(strings.TrimSpace(listOut.String()))

	for _, name := range names {
		if name == "." || name == ".." {
			continue
		}
		catCmd := exec.CommandContext(ctx, "docker", "exec", containerID, "sh", "-lc", "cat /workspace/"+shellQuote(name))
		var catOut, catErr bytes.Buffer
		catCmd.Stdout = &catOut
		catCmd.Stderr = &catErr
		if err := catCmd.Run(); err != nil {
			continue
		}
		destPath := workDir + string(os.PathSeparator) + name
		_ = os.WriteFile(destPath, catOut.Bytes(), 0o644)
	}
	return nil
}

func (d *DockerCLISandbox) createTaskContainer(ctx context.Context, req ExecRequest, command []string) (string, error) {
	seccompPath, cleanup, err := prepareSeccompProfile(ctx, SeccompProfileMode)
	if err != nil {
		return "", err
	}
	defer cleanup()

	args := []string{
		"create",
		"-i",
		"--network", "none",
		"--cpus", "1",
		"--memory", strconv.Itoa(req.MemoryLimitMB) + "m",
		"--pids-limit", "64",
		"--tmpfs", "/tmp:size=64m",
		"--cap-drop", "ALL",
		"--security-opt", "no-new-privileges",
		"--user", "1000:1000",
	}
	if d.transferMode() == "bind" {
		args = append(args, "--read-only")
		args = append(args, "-v", req.WorkDir+":/workspace")
	} else {
		// Docker Desktop / WSL2 环境下 tmpfs 默认 noexec，需要显式添加 exec 标志。
		args = append(args, "--tmpfs", "/workspace:size=64m,uid=1000,gid=1000,mode=0775,exec")
	}
	args = append(args, "-w", "/workspace")
	if seccompPath != "" {
		args = append(args, "--security-opt", "seccomp="+seccompPath)
	}
	args = append(args, req.Image)
	args = append(args, command...)

	cmd := exec.CommandContext(ctx, "docker", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", errors.New(strings.TrimSpace(nonEmpty(stderr.String(), err.Error())))
	}
	return strings.TrimSpace(stdout.String()), nil
}

func (d *DockerCLISandbox) prepareWorkspace(ctx context.Context, containerID string, req ExecRequest) error {
	if d.transferMode() == "bind" {
		return nil
	}
	if _, err := d.execInContainer(ctx, containerID, []string{"sh", "-lc", "mkdir -p /workspace"}, ""); err != nil {
		return err
	}

	entries, err := os.ReadDir(req.WorkDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		content, err := os.ReadFile(req.WorkDir + string(os.PathSeparator) + entry.Name())
		if err != nil {
			return err
		}
		// 使用 install(1) 设置 0755 权限，确保编译产物可执行。
		// --cap-drop ALL 移除了 CAP_FOWNER，容器内 chmod(1) 会失败。
		cmd := exec.CommandContext(ctx, "docker", "exec", "-i", containerID, "sh", "-lc",
			"install -m 755 /dev/null "+shellQuote("/workspace/"+entry.Name())+" && cat > "+shellQuote("/workspace/"+entry.Name()))
		cmd.Stdin = bytes.NewReader(content)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return errors.New(strings.TrimSpace(nonEmpty(stderr.String(), err.Error())))
		}
	}
	return nil
}

func (d *DockerCLISandbox) startAttached(ctx context.Context, containerID, stdin string) (string, string, error) {
	args := []string{"start", "-a"}
	if stdin != "" {
		args = append(args, "-i")
	}
	args = append(args, containerID)
	cmd := exec.CommandContext(ctx, "docker", args...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return stdout.String(), stderr.String(), nil
	}
	if err != nil && cmd.ProcessState == nil {
		return stdout.String(), stderr.String(), errors.New(strings.TrimSpace(nonEmpty(stderr.String(), err.Error())))
	}
	return stdout.String(), stderr.String(), nil
}

func (d *DockerCLISandbox) startDetached(ctx context.Context, containerID string) error {
	cmd := exec.CommandContext(ctx, "docker", "start", containerID)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return errors.New(strings.TrimSpace(nonEmpty(stderr.String(), err.Error())))
	}
	return nil
}

func (d *DockerCLISandbox) execInContainer(ctx context.Context, containerID string, command []string, stdin string) (ExecResult, error) {
	args := []string{"exec"}
	if stdin != "" {
		args = append(args, "-i")
	}
	args = append(args, containerID)
	args = append(args, command...)
	cmd := exec.CommandContext(ctx, "docker", args...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	exitCode := 0
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}
	result := ExecResult{
		ExitCode:    exitCode,
		Stdout:      stdout.String(),
		Stderr:      stderr.String(),
		StdoutBytes: stdout.Len(),
		StderrBytes: stderr.Len(),
	}
	if exitCode >= 128 {
		if sig := syscall.Signal(exitCode - 128); sig > 0 {
			result.Signal = sig.String()
		}
	}
	if ctx.Err() == context.DeadlineExceeded {
		result.TimedOut = true
		result.ExitCode = 124
		return result, nil
	}
	if err != nil && cmd.ProcessState == nil {
		return result, errors.New(strings.TrimSpace(nonEmpty(stderr.String(), err.Error())))
	}
	return result, nil
}

type dockerInspect struct {
	State struct {
		ExitCode  int    `json:"ExitCode"`
		OOMKilled bool   `json:"OOMKilled"`
		Error     string `json:"Error"`
	} `json:"State"`
}

func (d *DockerCLISandbox) inspectContainer(ctx context.Context, containerID string) (dockerInspect, error) {
	cmd := exec.CommandContext(ctx, "docker", "inspect", containerID)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return dockerInspect{}, errors.New(strings.TrimSpace(nonEmpty(stderr.String(), err.Error())))
	}
	var items []dockerInspect
	if err := json.Unmarshal(stdout.Bytes(), &items); err != nil {
		return dockerInspect{}, err
	}
	if len(items) == 0 {
		return dockerInspect{}, errors.New("empty inspect result")
	}
	return items[0], nil
}

func buildExecResult(inspect dockerInspect, runtime time.Duration, stdoutText, stderrText string, memoryKB int) ExecResult {
	result := ExecResult{
		ExitCode:    inspect.State.ExitCode,
		Stdout:      stdoutText,
		Stderr:      stderrText,
		Runtime:     runtime,
		MemoryKB:    memoryKB,
		StdoutBytes: len(stdoutText),
		StderrBytes: len(stderrText),
		OOMKilled:   inspect.State.OOMKilled,
	}
	if inspect.State.Error != "" && result.Stderr == "" {
		result.Stderr = inspect.State.Error
		result.StderrBytes = len(result.Stderr)
	}
	if inspect.State.ExitCode >= 128 {
		if sig := syscall.Signal(inspect.State.ExitCode - 128); sig > 0 {
			result.Signal = sig.String()
		}
	}
	return result
}

// readCgroupMemoryPeak 从容器内读取 cgroup v2 的 memory.peak。
// 返回 KB。若 cgroup v2 不可用则回退到 v1 的 max_usage_in_bytes。
func readCgroupMemoryPeak(ctx context.Context, containerID string) (int, error) {
	cmd := exec.CommandContext(ctx, "docker", "exec", containerID, "sh", "-lc",
		"cat /sys/fs/cgroup/memory.peak 2>/dev/null || cat /sys/fs/cgroup/memory/memory.max_usage_in_bytes 2>/dev/null")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return 0, errors.New(strings.TrimSpace(nonEmpty(stderr.String(), err.Error())))
	}
	line := strings.TrimSpace(stdout.String())
	if line == "" || line == "max" {
		return 0, nil
	}
	bytes, err := strconv.ParseInt(line, 10, 64)
	if err != nil {
		return 0, err
	}
	return int(bytes / 1024), nil
}

// wrapWithTimeMem wraps a command with /usr/bin/time -v to capture the
// per-process peak RSS (Maximum resident set size).  Unlike cgroup v2
// memory.peak (which is a read-only monotonic counter unsuitable for
// long-lived pooled containers), the per-process RSS is independent of
// container history and gives an accurate number every run.
//
// The wrapper writes a raw time log to memPeakPath+".raw" and then uses
// awk to extract "Maximum resident set size (kbytes)" × 1024 → bytes,
// writing the result to memPeakPath.  The original exit code is preserved.
func wrapWithTimeMem(command []string, memPeakPath string) []string {
	rawFile := memPeakPath + ".raw"
	// time -v -o RAW CMD ; awk RSS*1024 > OUT ; exit with CMD's code
	suffix := "; _rc=$?; awk '/Maximum resident/{print $NF*1024}' " +
		shellQuote(rawFile) + " > " + shellQuote(memPeakPath) +
		" 2>/dev/null; exit $_rc"

	if len(command) == 3 && command[0] == "sh" && command[1] == "-lc" {
		// Already a sh -lc command.  Nest another sh -lc inside time
		// so that env-var assignments (e.g. GOCACHE=/tmp) are parsed
		// by the inner shell, not by time itself.
		cmd := "time -v -o " + shellQuote(rawFile) + " sh -lc " + shellQuote(command[2]) + suffix
		return []string{"sh", "-lc", cmd}
	}
	// Argv command: wrap in sh -lc.
	parts := make([]string, len(command))
	for i, arg := range command {
		parts[i] = shellQuote(arg)
	}
	cmd := "time -v -o " + shellQuote(rawFile) + " " + strings.Join(parts, " ") + suffix
	return []string{"sh", "-lc", cmd}
}

// wrapWithMemPeak is a legacy wrapper that reads cgroup memory.peak.
// Prefer wrapWithTimeMem for per-process RSS accuracy in pooled containers.
func wrapWithMemPeak(command []string, memPeakPath string) []string {
	return wrapWithTimeMem(command, memPeakPath)
}

// removeContainer 移除临时容器。
func (d *DockerCLISandbox) removeContainer(ctx context.Context, containerID string) {
	_ = exec.CommandContext(ctx, "docker", "rm", "-f", containerID).Run()
}

func (d *DockerCLISandbox) transferMode() string {
	if d.TransferMode == "" {
		return "bind"
	}
	return d.TransferMode
}

// ensureImage 确保目标镜像在本地存在。
func (d *DockerCLISandbox) ensureImage(ctx context.Context, image string) error {
	if image == "" {
		return errors.New("empty image")
	}
	if _, ok := imageReady.Load(image); ok {
		return nil
	}

	inspectCmd := exec.CommandContext(ctx, "docker", "image", "inspect", image)
	if err := inspectCmd.Run(); err == nil {
		imageReady.Store(image, struct{}{})
		return nil
	}

	pullCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	pullCmd := exec.CommandContext(pullCtx, "docker", "pull", image)
	if output, err := pullCmd.CombinedOutput(); err != nil {
		return errors.New(string(output))
	}
	imageReady.Store(image, struct{}{})
	return nil
}

// nonEmpty 返回第一个非空字符串。
func nonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}
