func configCmd() {
if len(os.Args) < 3 {
figHelp()

}

subcommand := os.Args[2]
configPath := getConfigPath()
userSettingsPath := filepath.Join(filepath.Dir(configPath), "usersetting.json")

switch subcommand {
case "show":
loadConfig()
json.MarshalIndent(cfg, "", "  ")
tf("Current Configuration (including usersetting.json):\n%s\n", string(data))
case "set":
(os.Args) < 5 {
tln("Usage: marubot config set <key> <value>")

os.Args[3]
settings map[string]interface{}
os.ReadFile(userSettingsPath)
nil {
.Unmarshal(data, &settings)
gs = make(map[string]interface{})
to parse as JSON if it looks like one, otherwise keep as string
terface{}
json.Unmarshal([]byte(value), &val); err != nil {
// stay as string
gs[key] = val

ewData, _ := json.MarshalIndent(settings, "", "  ")
gsPath, newData, 0644)
tf("✓ Saved '%s' = %s to %s\n", key, value, userSettingsPath)
case "reset":
gsPath)
tln("✓ User settings reset to defaults.")
default:
figHelp()
}
}

func configHelp() {
fmt.Println("\nConfig commands:")
fmt.Println("  reset             Remove all user overrides")
fmt.Println("  set <key> <val>   Set an override in usersetting.json")
fmt.Println("  show              Show merged configuration")
}

func statusCmd() {
cfg, err := loadConfig()
if err != nil {
tf("Error loading config: %v\n", err)

}

configPath := getConfigPath()

fmt.Printf("%s marubot Status\n\n", logo)

if _, err := os.Stat(configPath); err == nil {
tln("Config:", configPath, "✓")
} else {
tln("Config:", configPath, "✗")
}

workspace := cfg.WorkspacePath()
if _, err := os.Stat(workspace); err == nil {
tln("Workspace:", workspace, "✓")
} else {
tln("Workspace:", workspace, "✗")
}

if _, err := os.Stat(configPath); err == nil {
tf("Model: %s\n", cfg.Agents.Defaults.Model)

gsPath := filepath.Join(filepath.Dir(configPath), "usersetting.json")
:= os.Stat(userSettingsPath); err == nil {
tln("User Settings:", userSettingsPath, "✓")
Router := cfg.Providers.OpenRouter.APIKey != ""
thropic := cfg.Providers.Anthropic.APIKey != ""
AI := cfg.Providers.OpenAI.APIKey != ""
i := cfg.Providers.Gemini.APIKey != ""
 != ""
 := cfg.Providers.Groq.APIKey != ""
!= ""

func(key string) string {
== "" { return "not set" }
(key) <= 8 { return "✓ (set)" }
 fmt.Sprintf("✓ %s...%s", key[:4], key[len(key)-4:])
c(enabled bool) string {
abled { return "✓" }
 "not set"
tln("OpenRouter API:", status(hasOpenRouter))
tln("Anthropic API:", status(hasAnthropic))
tln("OpenAI API:", status(hasOpenAI))
tln("Gemini API:", status(hasGemini))
tln("Zhipu API:", status(hasZhipu))
tln("Groq API:", status(hasGroq))
tf("vLLM/Local API: ✓\n")
tf("  - Base: %s\n", cfg.Providers.VLLM.APIBase)
tf("  - Key:  %s\n", maskKey(cfg.Providers.VLLM.APIKey))
tln("vLLM/Local: not set")
c getResourceDir() string {
home, _ := os.UserHomeDir()
return filepath.Join(home, ".marubot")
}

func getConfigPath() string {
return filepath.Join(getResourceDir(), "config.json")
}

func installAndRunSystemdService(exePath string) error {
_, err := exec.LookPath(\"systemctl\")
if err != nil {
 fmt.Errorf(\"systemctl not found\")
}

u, err := user.Current()
if err != nil {
 err
}

serviceDir := filepath.Join(u.HomeDir, \".config\", \"systemd\", \"user\")
if err := os.MkdirAll(serviceDir, 0755); err != nil {
 err
}

servicePath := filepath.Join(serviceDir, \"marubot.service\")
serviceContent := fmt.Sprintf(`[Unit]
Description=MaruBot Service
After=network.target

[Service]
Type=simple
Environment=\"MARUBOT_DAEMON=1\"
ExecStart=%s start --foreground
Restart=always
RestartSec=3

[Install]
WantedBy=default.target
`, exePath)

if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
 err
}

exec.Command(\"loginctl\", \"enable-linger\", u.Username).Run()

cmds := [][]string{
\"daemon-reload\"},
\"enable\", \"marubot.service\"},
\"start\", \"marubot.service\"},
}

for _, c := range cmds {
d(c[0], c[1:]...)
v(\"XDG_RUNTIME_DIR\") == \"\" {
v = append(os.Environ(), fmt.Sprintf(\"XDG_RUNTIME_DIR=/run/user/%s\", u.Uid))
v = os.Environ()
() // ignore specific errors
}

return nil
}

func reloadCmd() {
fmt.Println(\"Reloading MaruBot...\")
if runtime.GOOS == \"linux\" {
exec.LookPath(\"systemctl\")
nil {
user.Current()
u != nil {
:= \"\"
nil {
(u.HomeDir, \".config\", \"systemd\", \"user\")
(serviceDir, \"marubot.service\")
:= os.Stat(servicePath); err == nil {
d(\"systemctl\", \"--user\", \"restart\", \"marubot.service\")
v(\"XDG_RUNTIME_DIR\") == \"\" && uid != \"\" {
v = append(os.Environ(), fmt.Sprintf(\"XDG_RUNTIME_DIR=/run/user/%s\", uid))
v = os.Environ()
cmd.Run(); err == nil {
tln(\"✓ Reloaded via systemd.\")

time.Second)

exe, err := os.Executable()
if err != nil {
tf(\"✗ Executable path error: %v\n\", err)

}
cmd := exec.Command(exe, \"start\")
if err := cmd.Start(); err != nil {
tf(\"✗ Failed to start during reload: %v\n\", err)

}
fmt.Println(\"✓ Reload complete.\")
}

func startCmd() {
// Check for flags
var runForeground bool
if len(os.Args) > 2 && (os.Args[2] == \"--foreground\" || os.Args[2] == \"-f\") {
Foreground = true
}

// Double-fork / Detach logic
if !runForeground && os.Getenv(\"MARUBOT_DAEMON\") != \"1\" {
os.Executable()
nil {
tf(\"Error getting executable path: %v\n\", err)

time.GOOS == \"linux\" {
stallAndRunSystemdService(exe)
nil {
tln(\"✨ MaruBot started as a systemd service.\")
tln(\"   It will auto-restart on reboot and continue working.\")
tln(\"   URL: http://localhost:8080\")
tln(\"   To stop: use 'marubot stop'\")
tln(\"   To reload config: use 'marubot reload'\")

tf(\"Systemd service setup failed: %v. Falling back to simple daemon...\n\", err)
 with special env var
d(exe, \"start\")
v = append(os.Environ(), \"MARUBOT_DAEMON=1\")
 = nil
il
il

cmd.Start(); err != nil {
tf(\"Failed to start background process: %v\n\", err)
tf(\"%d\", cmd.Process.Pid)), 0644)

tf(\"✨ MaruBot Dashboard started in background (PID: %d)\n\", cmd.Process.Pid)
tln(\"   URL: http://localhost:8080\")
tln(\"   To stop: use 'marubot stop'\")
tln(\"   To reload config: use 'marubot reload'\")
tln(\"   Logs: ~/.marubot/dashboard.log\")

}

// Daemon Code Starts Here

// Setup logging to file if daemon
if os.Getenv(\"MARUBOT_DAEMON\") == \"1\" {
(getResourceDir(), \"dashboard.log\")
os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
nil {
to log file
ote: This only redirects Go's fmt.Print output if we assign it,
true redirection we'd need syscalls which are OS specific.
just use a logger.
 in this cross-platform Go app without syscalls:
just let it run. Stdout/Stderr are discarded by the parent anyway.
g a log file is good practice.
runForeground {
tf(\"%s Starting MaruBot Dashboard & API Server...\n\", logo)
}

cfg, err := loadConfig()
if err != nil {
Foreground {
tf(\"Error loading config: %v\n\", err)

}

provider, err := providers.CreateProvider(cfg)
if err != nil {
Foreground {
tf(\"Error creating provider: %v\n\", err)
bus.NewMessageBus()

// Enable logging to file for background service
logFile := filepath.Join(getResourceDir(), \"dashboard.log\")
if err := logger.EnableFileLogging(logFile); err != nil {
Foreground {
tf(\"Warning: Failed to enable file logging: %v\n\", err)
tLoop := agent.NewAgentLoop(cfg, bus, provider, version)

gpioService := gpio.NewGPIOService(cfg, bus)
gpioService.Start(context.Background())
defer gpioService.Stop()

// Background Services
cronStorePath := filepath.Join(filepath.Dir(getConfigPath()), \"cron\", \"jobs.json\")
cronService := cron.NewCronService(cronStorePath, func(job *cron.CronJob) (string, error) {
 agentLoop.ProcessDirect(context.Background(), job.Payload.Message, \"cron:\"+job.ID)
})

heartbeatService := heartbeat.NewHeartbeatService(
c(msg string) (string, error) {
 agentLoop.ProcessDirect(context.Background(), msg, \"heartbeat\")
cel := context.WithCancel(context.Background())
defer cancel()

if err := cronService.Start(); err != nil && runForeground {
tf(\"Error starting cron service: %v\n\", err)
}
if err := heartbeatService.Start(); err != nil && runForeground {
tf(\"Error starting heartbeat service: %v\n\", err)
}
go agentLoop.Run(ctx)

channelManager, err := channels.NewManager(cfg, bus)
if err == nil {
channelManager.StartAll(ctx); err != nil && runForeground {
tf(\"Error starting channels: %v\n\", err)
Foreground {
tln(\"✓ Background services started (Cron, Heartbeat, Channels)\")
runForeground {
tf(\"Warning: Failed to initialize channel manager: %v\n\", err)
tln(\"✓ Background services started (Cron, Heartbeat)\")
itialize Dashboard Server
port := \"8080\"
server := dashboard.NewServer(\":\"+port, agentLoop, cfg, version)

if runForeground {
c() {
d)
tf(\"✓ Dashboard available at http://localhost:%s\n\", port)
server.Start(); err != nil {
Foreground {
tf(\"Error starting dashboard server: %v\n\", err)
c getPidFilePath() string {
return filepath.Join(getResourceDir(), \"marubot.pid\")
}

func stopCmd() {
stoppedViaSystemd := false
if runtime.GOOS == \"linux\" {
exec.LookPath(\"systemctl\")
nil {
user.Current()
u != nil {
:= \"\"
nil {
(u.HomeDir, \".config\", \"systemd\", \"user\")
(serviceDir, \"marubot.service\")
:= os.Stat(servicePath); err == nil {
d(\"systemctl\", \"--user\", \"stop\", \"marubot.service\")
v(\"XDG_RUNTIME_DIR\") == \"\" && uid != \"\" {
v = append(os.Environ(), fmt.Sprintf(\"XDG_RUNTIME_DIR=/run/user/%s\", uid))
v = os.Environ()
cmd.Run(); err == nil {
tln(\"✓ Stopped systemd service.\")
true
getPidFilePath()
data, err := os.ReadFile(pidFile)
if err != nil {
{
tln(\"No running marubot process found (pid file missing).\")

}

pidStr := strings.TrimSpace(string(data))
var pid int
fmt.Sscanf(pidStr, \"%d\", &pid)

proc, err := os.FindProcess(pid)
if err != nil {
{
tf(\"Process %d not found.\n\", pid)

}

fmt.Printf(\"Stopping marubot daemon (PID: %d)...\n\", pid)
if err := proc.Signal(os.Interrupt); err != nil {
tln(\"✓ Stopped.\")
}

func upgradeCmd() {
autoConfirm := false
if len(os.Args) > 2 && (os.Args[2] == \"--yes\") {
firm = true
}

fmt.Println(\"⚙️  Checking for updates...\")

latest, err := config.CheckLatestVersion()
if err != nil {
tf(\"⚠️  Failed to check latest version: %v\n\", err)
tln(\"Proceeding with forced upgrade...\")
} else {
fig.IsNewVersionAvailable(latest) && !autoConfirm {
tf(\"✅ You are already using the latest version (v%s).\n\", config.Version)
t(\"Do you want to reinstall anyway? [y/N]: \")
ewReader(os.Stdin)
se, _ := reader.ReadString('\n')
se = strings.TrimSpace(strings.ToLower(response))
se != \"y\" && response != \"yes\" {

config.IsNewVersionAvailable(latest) && !autoConfirm {
tf(\"✨ New version available: v%s (Current: v%s)\n\", latest, config.Version)
t(\"Do you want to upgrade? [Y/n]: \")
ewReader(os.Stdin)
se, _ := reader.ReadString('\n')
se = strings.TrimSpace(strings.ToLower(response))
se == \"n\" || response == \"no\" {

existing process if running
stopCmd()

fmt.Println(\"��� Upgrading MaruBot to the latest version...\")

// Use curl to download and run the install script
// We use the same install script as it handles updates gracefully (git pull if exists)
cmd := exec.Command(\"bash\", \"-c\", \"curl -fsSL https://raw.githubusercontent.com/dirmich/maru-bot/main/install.sh | bash\")

// Connect pipes to let user interact (for language selection, sudo password, etc.)
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr
cmd.Stdin = os.Stdin

if err := cmd.Run(); err != nil {
tf(\"❌ Upgrade failed: %v\n\", err)
tln(\"✨ Upgrade complete! Restarting MaruBot...\")
reloadCmd()
}
