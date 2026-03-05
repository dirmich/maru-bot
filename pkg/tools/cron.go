package tools

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dirmich/marubot/pkg/cron"
)

type CronTool struct {
	storePath string
}

func NewCronTool(storePath string) *CronTool {
	return &CronTool{
		storePath: storePath,
	}
}

func (t *CronTool) Name() string {
	return "cron"
}

func (t *CronTool) Description() string {
	return "Manage scheduled background jobs (add, list, remove) for the agent. Use this when the user asks you to remind them periodically or run a task daily/hourly. When adding a job, provide a clear task description in 'message' for your future self."
}

func (t *CronTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"add", "list", "remove"},
				"description": "Action to perform",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Short name for the job (required for add)",
			},
			"message": map[string]interface{}{
				"type":        "string",
				"description": "The prompt or task command you want to execute when the schedule triggers. This message will be sent back to you. (required for add)",
			},
			"schedule_type": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"cron", "every", "at"},
				"description": "Type of schedule (required for add)",
			},
			"schedule_expr": map[string]interface{}{
				"type":        "string",
				"description": "Schedule expression. For 'cron' use standard cron format (e.g. '0 9 * * *' for 9 AM). For 'every' use seconds (e.g. '3600'). For 'at' use unix timestamp milliseconds. (required for add)",
			},
			"job_id": map[string]interface{}{
				"type":        "string",
				"description": "ID of the job to remove (required for remove)",
			},
		},
		"required": []string{"action"},
	}
}

func (t *CronTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	action, ok := args["action"].(string)
	if !ok {
		return "", fmt.Errorf("action is required")
	}

	cs := cron.NewCronService(t.storePath, nil)
	if err := cs.Load(); err != nil {
		return "", fmt.Errorf("failed to load cron store: %v", err)
	}

	switch action {
	case "add":
		name, _ := args["name"].(string)
		message, _ := args["message"].(string)
		sType, _ := args["schedule_type"].(string)
		sExpr, _ := args["schedule_expr"].(string)

		if name == "" || message == "" || sType == "" || sExpr == "" {
			return "", fmt.Errorf("name, message, schedule_type, and schedule_expr are required for 'add'")
		}

		deliver := true // Default send output back to channel
		channel := ""
		to := ""

		if c, ok := ctx.Value(CtxKeyChannel).(string); ok && c != "cli" {
			channel = c
		}
		if chId, ok := ctx.Value(CtxKeyChatID).(string); ok {
			to = chId
		}

		var schedule cron.CronSchedule
		if sType == "every" {
			sec, err := strconv.ParseInt(sExpr, 10, 64)
			if err != nil {
				return "", fmt.Errorf("invalid every expression (must be seconds): %v", err)
			}
			ms := sec * 1000
			schedule = cron.CronSchedule{Kind: "every", EveryMS: &ms}
		} else if sType == "at" {
			ts, err := strconv.ParseInt(sExpr, 10, 64)
			if err != nil {
				return "", fmt.Errorf("invalid at expression (must be ms timestamp): %v", err)
			}
			schedule = cron.CronSchedule{Kind: "at", AtMS: &ts}
		} else {
			schedule = cron.CronSchedule{Kind: "cron", Expr: sExpr}
		}

		job, err := cs.AddJob(name, schedule, message, deliver, channel, to)
		if err != nil {
			return "", fmt.Errorf("failed to add job: %v", err)
		}
		return fmt.Sprintf("Successfully added job '%s' with ID %s\nDeliver to channel: '%s', Recipient: '%s'", job.Name, job.ID, channel, to), nil

	case "list":
		jobs := cs.ListJobs(false)
		if len(jobs) == 0 {
			return "No active cron jobs.", nil
		}
		res := "Active Cron Jobs:\n"
		for _, j := range jobs {
			scheduleDesc := ""
			if j.Schedule.Kind == "cron" {
				scheduleDesc = "cron: " + j.Schedule.Expr
			} else if j.Schedule.Kind == "every" {
				scheduleDesc = fmt.Sprintf("every %d seconds", *j.Schedule.EveryMS/1000)
			} else if j.Schedule.Kind == "at" {
				scheduleDesc = fmt.Sprintf("at timestamp %d", *j.Schedule.AtMS)
			}
			res += fmt.Sprintf("- ID: %s | Name: %s | Schedule: %s | Delivered to: %s\n  Message: %s\n\n", j.ID, j.Name, scheduleDesc, j.Payload.Channel, j.Payload.Message)
		}
		return res, nil

	case "remove":
		jobID, ok := args["job_id"].(string)
		if !ok || jobID == "" {
			return "", fmt.Errorf("job_id is required for 'remove'")
		}
		if cs.RemoveJob(jobID) {
			return fmt.Sprintf("Successfully removed job %s", jobID), nil
		}
		return fmt.Sprintf("Job %s not found", jobID), nil

	default:
		return "", fmt.Errorf("unknown action: %s", action)
	}
}
